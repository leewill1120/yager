package nfs

import (
	"fmt"
	"leewill1120/yager/drivers/volume"
	"os"
	"os/exec"

	//log "github.com/Sirupsen/logrus"
)

type Volume struct {
	volume.CommonVolume
	RemotePath  string
	StoreServIP string
}

func NewVolume(volumeName string, rspBody map[string]interface{}) (*Volume, error) {
	if err := checkFields(rspBody); err != nil {
		return nil, err
	}

	v := new(Volume)
	v.Name = volumeName
	v.Type = rspBody["type"].(string)
	v.StoreServIP = rspBody["host"].(string)
	v.RemotePath = rspBody["remotePath"].(string)
	v.Status = volume.INIT

	return v, nil
}

func (v *Volume) Attribute() map[string]interface{} {
	return map[string]interface{}{
		"Name":        v.Name,
		"Type":        v.Type,
		"Status":      v.Status,
		"StoreServIP": v.StoreServIP,
		"MountPoint":  v.MountPoint,
		"RemotePath":  v.RemotePath,
	}
}

func (v *Volume) Mount() error {
	mountPoint := "/mnt/yager/nfs/" + v.Name
	if err := os.MkdirAll(mountPoint, os.ModeDir); err != nil {
		return err
	}
	cmd := exec.Command("mount", v.StoreServIP+":"+v.RemotePath, mountPoint)
	if d, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf(string(d))
	}
	v.MountPoint = mountPoint
	v.Status = volume.MOUNTED

	return nil
}

func (v *Volume) Umount() error {
	cmd := exec.Command("umount", v.MountPoint)
	if d, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf(string(d))
	}
	v.Status = volume.UNMOUNTED
	return nil
}

func checkFields(rspBody map[string]interface{}) error {
	fields := []string{"type", "host", "remotePath"}
	for _, field := range fields {
		if _, exist := rspBody[field]; !exist {
			return fmt.Errorf("field(%s) not found.", field)
		}
	}
	return nil
}
