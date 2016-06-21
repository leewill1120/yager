package volume

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
)

const (
	INIT = iota
	OK
	BAD
)

type Volume struct {
	Name          string
	MountPoint    string
	Dev           string
	Target        string
	StoreServIP   string
	StoreServPort int
	UserID        string
	Password      string
	Size          float64
	Status        int
	Type          string //iscsi nfs cifs
}

func NewVolume(volumeName string, size float64, volumetype string) *Volume {
	return &Volume{
		Name:   volumeName,
		Size:   size,
		Status: INIT,
		Type:   volumetype,
	}
}

func (v *Volume) GetBackendStore(storeMServIP string, storeMServPort int, initiatorName string) error {
	var (
		err     error
		buf     []byte                 = make([]byte, 1024)
		reqBody map[string]interface{} = make(map[string]interface{})
		rspBody map[string]interface{} = make(map[string]interface{})
		url     string                 = "http://" + storeMServIP + ":" + strconv.Itoa(storeMServPort) + "/block/create"
		rsp     *http.Response
	)

	reqBody["type"] = v.Type
	reqBody["initiatorName"] = initiatorName
	reqBody["size"] = v.Size
	if buf, err = json.Marshal(reqBody); err != nil {
		return err
	}
	if rsp, err = http.Post(url, "application/json", bytes.NewBuffer(buf)); err != nil {
		return err
	}
	if 4 == rsp.StatusCode/100 || 5 == rsp.StatusCode/100 {
		return fmt.Errorf("server return %d.\n", rsp.StatusCode)
	}
	if buf, err = ioutil.ReadAll(rsp.Body); err != nil {
		return err
	}

	if err = json.Unmarshal(buf, &rspBody); err != nil {
		return err
	}

	if "success" != rspBody["result"].(string) {
		return fmt.Errorf("GetBackendStore failed.reason:%s\n", rspBody["detail"])
	}

	v.StoreServIP = rspBody["host"].(string)
	v.StoreServPort = int(rspBody["port"].(float64))
	v.UserID = rspBody["userid"].(string)
	v.Password = rspBody["password"].(string)
	v.Target = rspBody["target"].(string)
	v.Type = rspBody["type"].(string)

	return nil
}

func (v *Volume) ReleaseBackendStore(storeMServIP string, storeMServPort int) error {
	var (
		err     error
		buf     []byte                 = make([]byte, 1024)
		reqBody map[string]interface{} = make(map[string]interface{})
		rspBody map[string]interface{} = make(map[string]interface{})
		url     string                 = "http://" + storeMServIP + ":" + strconv.Itoa(storeMServPort) + "/block/delete"
		rsp     *http.Response
	)

	reqBody["target"] = v.Target
	if buf, err = json.Marshal(reqBody); err != nil {
		return err
	}
	if rsp, err = http.Post(url, "application/json", bytes.NewBuffer(buf)); err != nil {
		return err
	}
	if 4 == rsp.StatusCode/100 || 5 == rsp.StatusCode/100 {
		return fmt.Errorf("server return %d.\n", rsp.StatusCode)
	}
	if buf, err = ioutil.ReadAll(rsp.Body); err != nil {
		return err
	}

	if err = json.Unmarshal(buf, &rspBody); err != nil {
		return err
	}

	if "success" != rspBody["result"].(string) {
		return fmt.Errorf("ReleaseBackendStore failed.reason:%s\n", rspBody["detail"])
	}
	v.Status = BAD
	return nil
}

func (v *Volume) Umount() error {
	mountPoint := "/mnt/yager/" + v.Name
	cmd := exec.Command("umount", mountPoint)
	if d, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf(string(d))
	}
	return nil
}

func (v *Volume) FormatAndMount() error {
	cmd := exec.Command("mkfs.xfs", v.Dev)
	if d, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf(string(d))
	}

	mountPoint := "/mnt/yager/" + v.Name
	if err := os.MkdirAll(mountPoint, os.ModeDir); err != nil {
		return err
	}
	cmd = exec.Command("mount", v.Dev, mountPoint)
	if d, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf(string(d))
	}
	v.MountPoint = mountPoint
	return nil
}

func (v *Volume) LoginTarget() error {
	cmd := exec.Command("iscsiadm", "-m", "discovery", "-t", "sendtargets", "-p", v.StoreServIP)
	if d, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf(string(d))
	}

	cmd = exec.Command("iscsiadm", "-m", "node", "-T", v.Target, "-o", "update", "--name", "node.session.auth.authmethod", "--value=CHAP")
	if d, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf(string(d))
	}

	cmd = exec.Command("iscsiadm", "-m", "node", "-T", v.Target, "-o", "update", "--name", "node.session.auth.username", "--value="+v.UserID)
	if d, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf(string(d))
	}

	cmd = exec.Command("iscsiadm", "-m", "node", "-T", v.Target, "-o", "update", "--name", "node.session.auth.password", "--value="+v.Password)
	if d, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf(string(d))
	}

	cmd = exec.Command("iscsiadm", "-m", "node", "-T", v.Target, "-p", v.StoreServIP, "--login")
	if d, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf(string(d))
	}
	return nil
}

func (v *Volume) LogoutTarget() error {
	cmd := exec.Command("iscsiadm", "-m", "node", "-T", v.Target, "--logout")
	if d, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf(string(d))
	}
	return nil
}

func (v *Volume) LoginAndGetNewDev() error {
	p1 := getPartitions()

	if e := v.LoginTarget(); e != nil {
		return e
	}

	time.Sleep(time.Millisecond * 100)
	p2 := getPartitions()
	var newParition string
	for index, _ := range p2 {
		if _, exist := p1[index]; !exist {
			newParition = index
		}
	}
	if "" == newParition {
		return fmt.Errorf("new dev not found.")
	}
	v.Dev = "/dev/" + newParition
	return nil
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
