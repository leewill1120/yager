package lvm

import (
	"log"
	"os/exec"
	"strconv"
	"strings"

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

func NewVG() *VolumeGroup {
	vg := &VolumeGroup{
		Name: defaultVolumeGroup,
	}
	vg.update()
	return vg
}

func (vg *VolumeGroup) update() {
	vg.getVGSize()
	vg.getLVs()
}

//Get volume group total size and free size
func (vg *VolumeGroup) getVGSize() {
	cmd := exec.Command("vgs", vg.Name, "--units", "M")
	if data, err := cmd.Output(); err != nil {
		log.Fatal(err)
	} else {
		line := strings.Fields(strings.Split(string(data), "\n")[1])
		if number, err := strconv.ParseFloat(line[5][:len(line[5])-1], 64); err != nil {
			log.Fatal(err)
		} else {
			vg.Size = number
		}

		if number, err := strconv.ParseFloat(line[6][:len(line[6])-1], 64); err != nil {
			log.Fatal(err)
		} else {
			vg.Free = number
		}
	}
}

//Get logic volume belong to vg
func (vg *VolumeGroup) getLVs() {
	cmd := exec.Command("lvs", vg.Name, "--units", "M")
	if data, err := cmd.Output(); err != nil {
		log.Fatal(err)
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
				log.Fatal(err)
			} else {
				lv.Size = number
			}
			vg.Lvs = append(vg.Lvs, lv)
		}
	}
}

func (vg *VolumeGroup) CreateLV(size float64) (string, error) {
	name, _ := utils.Generate_wwn("unit_serail")
	cmd := exec.Command("lvcreate", vg.Name, "-L", strconv.FormatFloat(size, 'f', -1, 64)+"M", "-n", name)

	stderr, _ := cmd.StderrPipe()
	defer stderr.Close()
	go func() {
		for {
			p := make([]byte, 1024)
			if length, err := stderr.Read(p); err == nil {
				log.Println(string(p[:length-1]))
			} else {
				break
			}
		}
	}()

	if data, err := cmd.Output(); err != nil {
		return "", err
	} else {
		vg.update()
		log.Println(string(data))
		return name, nil
	}
}

func (vg *VolumeGroup) RemoveLV(lvname string) error {
	for _, lv := range vg.Lvs {
		if lv.Name == lvname {
			lvpath := "/dev/" + vg.Name + "/" + lvname
			cmd := exec.Command("lvremove", lvpath, "-f")
			stderr, _ := cmd.StderrPipe()
			defer stderr.Close()
			go func() {
				for {
					p := make([]byte, 1024)
					if length, err := stderr.Read(p); err == nil {
						log.Println(string(p[:length-1]))
					} else {
						break
					}
				}
			}()

			if _, err := cmd.Output(); err != nil {
				return err
			} else {
				vg.update()
				return nil
			}
		}
	}
	log.Printf("Logic volume(%s) doesn't exists.", lvname)
	return nil
}

func (vg *VolumeGroup) RemoveAllLV() error {
	cmd := exec.Command("lvremove", vg.Name, "-f")
	stderr, _ := cmd.StderrPipe()
	defer stderr.Close()
	go func() {
		for {
			p := make([]byte, 1024)
			if length, err := stderr.Read(p); err == nil {
				log.Println(string(p[:length-1]))
			} else {
				break
			}
		}
	}()

	if _, err := cmd.Output(); err != nil {
		return err
	} else {
		vg.update()
		return nil
	}
}
