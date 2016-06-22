package iscsiadm

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	"leewill1120/yager/drivers/volume"

	log "github.com/Sirupsen/logrus"
)

type Volume struct {
	volume.CommonVolume
	Dev           string
	Target        string
	StoreServIP   string
	StoreServPort int
	UserID        string
	Password      string
	Size          float64
	Formatted     bool
}

func checkFields(rspBody map[string]interface{}) error {
	fields := []string{"type", "host", "port", "size", "target", "userid", "password"}
	for _, field := range fields {
		if _, exist := rspBody[field]; !exist {
			return fmt.Errorf("field(%s) not found.", field)
		}
	}
	return nil
}

func NewVolume(volumeName string, rspBody map[string]interface{}) (*Volume, error) {
	if err := checkFields(rspBody); err != nil {
		return nil, err
	}

	v := new(Volume)
	v.Name = volumeName
	v.Type = rspBody["type"].(string)
	v.StoreServIP = rspBody["host"].(string)
	v.StoreServPort = int(rspBody["port"].(float64))
	v.Size = rspBody["size"].(float64)
	v.Target = rspBody["target"].(string)
	v.UserID = rspBody["userid"].(string)
	v.Password = rspBody["password"].(string)
	v.Status = volume.INIT
	v.Formatted = false

	return v, nil
}

func (v *Volume) Attribute() map[string]interface{} {
	return map[string]interface{}{
		"Name":          v.Name,
		"Type":          v.Type,
		"Target":        v.Target,
		"Status":        v.Status,
		"Size":          v.Size,
		"StoreServIP":   v.StoreServIP,
		"StoreServPort": v.StoreServPort,
		"UserID":        v.UserID,
		"Password":      v.Password,
		"Dev":           v.Dev,
		"MountPoint":    v.MountPoint,
		"Formatted":     v.Formatted,
	}
}

func (v *Volume) Mount() error {
	//1.LoginTarget
	//2.GetAddedDiv
	//3.Format & Mount

	if err := v.loginTarget(); err != nil {
		log.Debug(err)
		return err
	}

	if err := v.getAddedDev(); err != nil {
		log.Debug(err)
		return err
	}

	if err := v.formatAndMount(); err != nil {
		log.Debug(err)
		return err
	}
	v.Status = volume.MOUNTED

	return nil
}

func (v *Volume) Umount() error {
	mountPoint := "/mnt/yager/" + v.Name
	cmd := exec.Command("umount", mountPoint)
	if d, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf(string(d))
	}
	v.Status = volume.UNMOUNTED

	return v.logoutTarget()
}

func (v *Volume) getAddedDev() error {
	p1 := getPartitions()

	time.Sleep(time.Millisecond * 100)
	p2 := getPartitions()
	var newParition string
	for index, _ := range p2 {
		if _, exist := p1[index]; !exist {
			newParition = index
		}
	}
	if "" == newParition {
		return fmt.Errorf("no new added device.")
	}
	v.Dev = "/dev/" + newParition
	return nil
}

func (v *Volume) formatAndMount() error {
	if !v.Formatted {
		cmd := exec.Command("mkfs.xfs", v.Dev)
		if d, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf(string(d))
		}
		v.Formatted = true
	}

	mountPoint := "/mnt/yager/" + v.Name
	if err := os.MkdirAll(mountPoint, os.ModeDir); err != nil {
		return err
	}
	cmd := exec.Command("mount", v.Dev, mountPoint)
	if d, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf(string(d))
	}
	v.MountPoint = mountPoint
	return nil
}

func (v *Volume) loginTarget() error {
	cmd := exec.Command("iscsiadm", "-m", "discovery", "-t", "sendtargets", "-p", v.StoreServIP)
	if d, err := cmd.CombinedOutput(); err != nil {
		log.Debug(string(d))
		return fmt.Errorf(string(d))
	}

	cmd = exec.Command("iscsiadm", "-m", "node", "-T", v.Target, "-o", "update", "--name", "node.session.auth.authmethod", "--value=CHAP")
	if d, err := cmd.CombinedOutput(); err != nil {
		log.Debug(string(d))
		return fmt.Errorf(string(d))
	}

	cmd = exec.Command("iscsiadm", "-m", "node", "-T", v.Target, "-o", "update", "--name", "node.session.auth.username", "--value="+v.UserID)
	if d, err := cmd.CombinedOutput(); err != nil {
		log.Debug(string(d))
		return fmt.Errorf(string(d))
	}

	cmd = exec.Command("iscsiadm", "-m", "node", "-T", v.Target, "-o", "update", "--name", "node.session.auth.password", "--value="+v.Password)
	if d, err := cmd.CombinedOutput(); err != nil {
		log.Debug(string(d))
		return fmt.Errorf(string(d))
	}

	cmd = exec.Command("iscsiadm", "-m", "node", "-T", v.Target, "-p", v.StoreServIP, "--login")
	if d, err := cmd.CombinedOutput(); err != nil {
		log.Debug(string(d))
		return fmt.Errorf(string(d))
	}
	return nil
}

func (v *Volume) logoutTarget() error {
	cmd := exec.Command("iscsiadm", "-m", "node", "-T", v.Target, "--logout")
	if d, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf(string(d))
	}
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
