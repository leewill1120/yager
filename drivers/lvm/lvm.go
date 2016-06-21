package lvm

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"

	"leewill1120/yager/utils"
)

var (
	defaultVolumeGroup string = "yager"
)

type LogicVolume struct {
	Name string
	Size float64
}

type VolumeGroup struct {
	Name string
	Size float64 //MB
	Free float64 //MB
	Lvs  []LogicVolume
}

func NewVG(vgname string) *VolumeGroup {
	if "" == vgname {
		vgname = defaultVolumeGroup
	}
	vg := &VolumeGroup{
		Name: vgname,
	}
	if e := vg.update(); e != nil {
		log.Fatal(e)
	}
	return vg
}
func (vg *VolumeGroup) update() error {
	e1 := vg.getVGSize()
	e2 := vg.getLVs()
	if e1 == nil && e2 == nil {
		return nil
	} else {
		return fmt.Errorf("getVGSize: %s, getLVs:%s.", e1.Error(), e2.Error())
	}
}

//Get volume group total size and free size
func (vg *VolumeGroup) getVGSize() error {
	cmd := exec.Command("vgs", vg.Name, "--units", "M")
	if data, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf(strings.Replace(string(data), "\n", " ", -1))
	} else {
		line := strings.Fields(strings.Split(string(data), "\n")[1])
		if number, err := strconv.ParseFloat(line[5][:len(line[5])-1], 64); err != nil {
			return err
		} else {
			vg.Size = number
		}

		if number, err := strconv.ParseFloat(line[6][:len(line[6])-1], 64); err != nil {
			return err
		} else {
			vg.Free = number
		}
		return nil
	}
}

//Get logic volume belong to vg
func (vg *VolumeGroup) getLVs() error {
	cmd := exec.Command("lvs", vg.Name, "--units", "M")
	if data, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf(strings.Replace(string(data), "\n", " ", -1))
	} else {
		vg.Lvs = make([]LogicVolume, 0)
		lines := strings.Split(string(data), "\n")[1:]
		for _, line := range lines {
			if 0 == len(line) {
				continue
			}
			var lv LogicVolume
			lineSlice := strings.Fields(line)
			lv.Name = lineSlice[0]
			if number, err := strconv.ParseFloat(lineSlice[3][:len(lineSlice[3])-1], 64); err != nil {
				return err
			} else {
				lv.Size = number
			}
			vg.Lvs = append(vg.Lvs, lv)
		}
		return nil
	}
}

func (vg *VolumeGroup) CreateLV(size float64) (string, error) {
	if size > vg.Free {
		return "", fmt.Errorf("Volume group  has insufficient free space")
	}
	name, _ := utils.Generate_wwn("unit_serail")
	cmd := exec.Command("lvcreate", vg.Name, "-L", strconv.FormatFloat(size, 'f', -1, 64)+"M", "-n", name)

	if data, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf(strings.Replace(string(data), "\n", " ", -1))
	} else {
		vg.update()
		return name, nil
	}
}

func (vg *VolumeGroup) RemoveLV(lvname string) error {
	for _, lv := range vg.Lvs {
		if lv.Name == lvname {
			lvpath := "/dev/" + vg.Name + "/" + lvname
			cmd := exec.Command("lvremove", lvpath, "-f")

			if data, err := cmd.CombinedOutput(); err != nil {
				return fmt.Errorf(strings.Replace(string(data), "\n", " ", -1))
			} else {
				vg.update()
				return nil
			}
		}
	}
	log.WithFields(log.Fields{
		"logic volume": lvname,
	}).Warn("logic volume doesn't exists.")
	return nil
}

//remove all lv which belongs this vg.
func (vg *VolumeGroup) Clear() error {
	cmd := exec.Command("lvremove", vg.Name, "-f")

	if data, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf(strings.Replace(string(data), "\n", " ", -1))
	} else {
		vg.update()
		return nil
	}
}
