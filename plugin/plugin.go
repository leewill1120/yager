package plugin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"leewill1120/mux"
	"leewill1120/yager/plugin/volume"
)

var (
	defaultInitiatorNameFile string = "/etc/iscsi/initiatorname.iscsi"
	defaultUnixDomainSocket  string = "/run/docker/plugins/yager.sock"
)

type Block struct {
	IP       string
	Target   string
	Username string
	Password string
}

type Plugin struct {
	StoreServIP   string
	StoreServPort int
	BlockList     map[string]string
	VolumeList    map[string]*volume.Volume
	InitiatorName string
}

func NewPlugin(storeServIP string, storeServPort int) *Plugin {
	p := &Plugin{
		StoreServIP:   storeServIP,
		StoreServPort: storeServPort,
	}

	if d, e := ioutil.ReadFile(defaultInitiatorNameFile); e != nil {
		log.Fatal(e)
	} else {
		InitiatorName := strings.Replace(string(d), "\n", "", -1)
		InitiatorName = strings.Replace(string(d), "InitiatorName=", "", -1)
		InitiatorName = strings.TrimSpace(InitiatorName)
		p.InitiatorName = InitiatorName
	}
	if err := p.FromDisk(); err != nil {
		log.Fatal(err)
	}
	return p
}

func (p *Plugin) Run() {
	p.createUnixSock(defaultUnixDomainSocket)

	router := mux.NewRouter()
	router.HandleFunc("/Plugin.Activate", p.Activate).Methods("POST")
	router.HandleFunc("/VolumeDriver.Create", p.CreateVolume).Methods("POST")
	router.HandleFunc("/VolumeDriver.Remove", p.RemoveVolume).Methods("POST")
	router.HandleFunc("/VolumeDriver.Mount", p.MountVolume).Methods("POST")
	router.HandleFunc("/VolumeDriver.Path", p.VolumePath).Methods("POST")
	router.HandleFunc("/VolumeDriver.Get", p.GetVolume).Methods("POST")
	router.HandleFunc("/VolumeDriver.List", p.ListVolumes).Methods("POST")

	addr := &net.UnixAddr{
		Name: defaultUnixDomainSocket,
		Net:  "unix",
	}
	l, e := net.ListenUnix("unix", addr)
	if e != nil {
		log.Fatal(e)
	}
	http.Serve(l, router)
}

func (p *Plugin) createUnixSock(unixsock string) error {
	if _, err := os.Stat(unixsock); err != nil {
		os.MkdirAll(path.Dir(unixsock), os.ModeDir)
	}
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	go func() {
		fmt.Printf("Received stop signal(%s)\n", <-c)
		os.Remove(unixsock)
		os.Exit(0)
	}()
	return nil
}

func (p *Plugin) Activate(rsp http.ResponseWriter, req *http.Request) {
	rspBody := make(map[string]interface{})
	implements := make([]string, 0)
	implements = append(implements, "VolumeDriver")
	rspBody["Implements"] = implements
	if buf, err := json.Marshal(rspBody); err != nil {
		log.Println(err)
		rsp.Write([]byte(err.Error()))
	} else {
		_, err = rsp.Write(buf)
		if err != nil {
			log.Println(err)
		}
	}
}

func (p *Plugin) FromDisk() error {
	home := os.Getenv("HOME")
	configPath := home + "/.yager.json"
	if d, e := ioutil.ReadFile(configPath); e != nil {
		return e
	} else {
		if e = json.Unmarshal(d, &p.VolumeList); e != nil {
			log.Println(e)
			return e
		}
	}
	return nil
}

func (p *Plugin) ToDisk() {
	home := os.Getenv("HOME")
	configPath := home + "/.yager.json"
	if b, err := json.Marshal(p.VolumeList); err != nil {
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

/*
func (p *Plugin) CmdGetBlock(size float64) {
	if d, e := ioutil.ReadFile(defaultInitiatorNameFile); e != nil {
		log.Fatal(e)
	} else {
		InitiatorName := strings.Replace(string(d), "\n", "", -1)
		InitiatorName = strings.Replace(string(d), "InitiatorName=", "", -1)
		InitiatorName = strings.TrimSpace(InitiatorName)
		b := getBlock(InitiatorName, c.StoreServIP, c.StoreServPort, size)
		if b == nil {
			log.Fatal("failed to get block")
		}
		newParition := getNewPartition(b)
		c.BlockList["/dev/"+newParition] = b.Target
		c.saveBlockList()
		fmt.Printf("got new dev: /dev/%s\n", newParition)
	}
}
*/
