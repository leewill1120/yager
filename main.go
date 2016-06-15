package main

import (
	"fmt"
	"leewill1120/yager/utils"
)

func main() {
	fmt.Println(utils.Generate_wwn("iqn"))
	fmt.Println(utils.Generate_wwn("naa"))
	fmt.Println(utils.Generate_wwn("eui"))
	fmt.Println(utils.Generate_wwn("unit_serail"))
}
