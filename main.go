package main

import (
	"fmt"
	"log"

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
		fmt.Println("======================================================")
		if target, e := c.AddTarget("/backstores/liwei/lv1", "iqn.2016-06.org.baidu:server01", "user1", "passwd"); e != nil {
			log.Println(e)
		} else {
			c.Print()
			fmt.Println("======================================================")
			c.RemoveTarget(target)
			c.Print()
		}
	}
}
