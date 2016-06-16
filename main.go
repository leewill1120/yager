package main

import (
	"fmt"
	"log"
	"time"

	"leewill1120/yager/drivers/lvm"
	"leewill1120/yager/drivers/rtslib"
	"leewill1120/yager/utils"
)

func main() {
	utils.Generate_wwn("iqn")
	c := rtslib.NewConfig()
	if e := c.FromDisk(""); e != nil {
		fmt.Println(e)
	} else {
		c.Print()
		if target, e := c.AddTarget("/backstores/liwei/lv1", "iqn.2016-06.org.baidu:server01", "user1", "passwd"); e != nil {
			log.Println(e)
		} else {
			c.Print()
			c.RemoveTarget(target)
			c.Print()
		}
	}
	vg := lvm.NewVG()

	var lvs []string
	for i := 0; i < 10; i++ {
		lv, _ := vg.CreateLV(100)
		lvs = append(lvs, lv)
	}
	fmt.Println(vg)

	time.Sleep(time.Second * 1)
	for _, lv := range lvs {
		vg.RemoveLV(lv)
	}

	vg.RemoveAllLV()
	fmt.Println(vg)
	time.Sleep(time.Second * 1)
}
