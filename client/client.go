package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	defaultInitiatorNameFile string = "/etc/iscsi/initiatorname.iscsi"
)

type Block struct {
	IP       string
	Target   string
	Username string
	Password string
}

type Client struct {
	ServIP    string
	ServPort  string
	BlockList map[string]string
}

func NewClient(ip, port string) *Client {
	c := &Client{
		ServIP:    ip,
		ServPort:  port,
		BlockList: make(map[string]string),
	}
	c.loadBlockList()
	return c
}

func (c *Client) loadBlockList() {
	home := os.Getenv("HOME")
	configPath := home + "/.yager.json"
	if d, e := ioutil.ReadFile(configPath); e != nil {
		return
	} else {
		if e = json.Unmarshal(d, &c.BlockList); e != nil {
			log.Println(e)
		}
	}
}

func (c *Client) saveBlockList() {
	home := os.Getenv("HOME")
	configPath := home + "/.yager.json"
	if b, err := json.Marshal(&c.BlockList); err != nil {
		log.Println(err)
	} else {
		if err = ioutil.WriteFile(configPath, b, 0755); err != nil {
			log.Println(err)
		}
	}
}

func getPartitions() map[string]struct{} {
	Partitions := make(map[string]struct{})
	file := "/proc/partitions"
	if d, e := ioutil.ReadFile(file); e != nil {
		log.Fatal(e)
	} else {
		for _, line := range strings.Split(string(d), "\n")[1:] {
			if "" == line {
				continue
			}
			seg := strings.Fields(line)
			Partitions[seg[len(seg)-1]] = struct{}{}
		}
	}
	return Partitions
}

func getBlock(initiatorname, ip, port string, size float64) *Block {
	context := map[string]interface{}{
		"InitiatorName": initiatorname,
		"Size":          size,
	}
	bs, err := json.Marshal(context)
	if err != nil {
		return nil
	}
	body := bytes.NewBuffer(bs)
	if rsp, err := http.Post("http://"+ip+":"+port+"/block/create", "application/json", body); err != nil {
		fmt.Println(err)
		return nil
	} else {
		var rspMap map[string]interface{}
		jd := json.NewDecoder(rsp.Body)
		if err := jd.Decode(&rspMap); err != nil {
			fmt.Println(err)
			return nil
		} else {
			if "success" == (rspMap["result"]).(string) {
				return &Block{
					Target:   (rspMap["target"]).(string),
					Username: (rspMap["userid"]).(string),
					Password: (rspMap["password"]).(string),
					IP:       ip,
				}
			} else {
				fmt.Println(rspMap["detail"])
				return nil
			}
		}
	}
}

func removeBlock(target, ip, port string) error {
	context := map[string]interface{}{
		"target": target,
	}
	bs, err := json.Marshal(context)
	if err != nil {
		return err
	}
	body := bytes.NewBuffer(bs)
	if rsp, err := http.Post("http://"+ip+":"+port+"/block/delete", "application/json", body); err != nil {
		return err
	} else {
		var rspMap map[string]interface{}
		jd := json.NewDecoder(rsp.Body)
		if err := jd.Decode(&rspMap); err != nil {
			return err
		} else {
			if "success" == (rspMap["result"]).(string) {
				return nil
			} else {
				return fmt.Errorf((rspMap["detail"]).(string))
			}
		}
	}
}

//iscsiadm -m discovery -t sendtargets $ipaddr
func loginTarget(b *Block) error {
	cmd := exec.Command("iscsiadm", "-m", "discovery", "-t", "sendtargets", "-p", b.IP)
	if d, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf(string(d))
	}

	cmd = exec.Command("iscsiadm", "-m", "node", "-T", b.Target, "-o", "update", "--name", "node.session.auth.authmethod", "--value=CHAP")
	if d, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf(string(d))
	}

	cmd = exec.Command("iscsiadm", "-m", "node", "-T", b.Target, "-o", "update", "--name", "node.session.auth.username", "--value="+b.Username)
	if d, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf(string(d))
	}

	cmd = exec.Command("iscsiadm", "-m", "node", "-T", b.Target, "-o", "update", "--name", "node.session.auth.password", "--value="+b.Password)
	if d, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf(string(d))
	}

	cmd = exec.Command("iscsiadm", "-m", "node", "-T", b.Target, "-p", b.IP, "--login")
	if d, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf(string(d))
	}
	return nil
}

//iscsiadm -m node -T iqn.2016-06.org.linux-iscsi.iscsi001.amd64:sn.11bad46ba54b --logout
func logoutTarget(target string) error {
	cmd := exec.Command("iscsiadm", "-m", "node", "-T", target, "--logout")
	if d, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf(string(d))
	}
	return nil
}

func getNewPartition(b *Block) string {
	p1 := getPartitions()

	if e := loginTarget(b); e != nil {
		log.Fatal(e)
	}

	time.Sleep(time.Millisecond * 100)
	p2 := getPartitions()
	var newParition string
	for index, _ := range p2 {
		if _, exist := p1[index]; !exist {
			newParition = index
		}
	}
	return newParition
}

func (c *Client) CmdGetBlock(size float64) {
	if d, e := ioutil.ReadFile(defaultInitiatorNameFile); e != nil {
		log.Fatal(e)
	} else {
		InitiatorName := strings.Replace(string(d), "\n", "", -1)
		InitiatorName = strings.Replace(string(d), "InitiatorName=", "", -1)
		InitiatorName = strings.TrimSpace(InitiatorName)
		b := getBlock(InitiatorName, c.ServIP, c.ServPort, size)
		if b == nil {
			log.Fatal("failed to get block")
		}
		newParition := getNewPartition(b)
		c.BlockList["/dev/"+newParition] = b.Target
		c.saveBlockList()
		fmt.Printf("got new dev: /dev/%s\n", newParition)
	}
}

//c.CmdRemoveBlock
func (c *Client) CmdRemoveBlock(dev string) {
	target := c.BlockList[strings.TrimSpace(dev)]

	if err := logoutTarget(target); err != nil {
		log.Fatal(err)
	}

	if err := removeBlock(target, c.ServIP, c.ServPort); err != nil {
		log.Fatal(err)
	}

	delete(c.BlockList, strings.TrimSpace(dev))
	c.saveBlockList()
	fmt.Printf("removed block: %s\n", dev)
}
