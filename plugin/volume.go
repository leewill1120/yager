package plugin

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"leewill1120/yager/plugin/volume"
)

func (p *Plugin) CreateVolume(rsp http.ResponseWriter, req *http.Request) {
	var (
		err     error
		buf     []byte                 = make([]byte, 1024)
		reqBody map[string]interface{} = make(map[string]interface{})
		rspBody map[string]string      = make(map[string]string)
		size    float64
	)

	defer func() {
		if buf, err = json.Marshal(rspBody); err != nil {
			log.Println(err)
			return
		}
		if _, err = rsp.Write(buf); err != nil {
			log.Println(err)
			return
		}
	}()

	if buf, err = ioutil.ReadAll(req.Body); err != nil {
		rspBody["Err"] = err.Error()
		return
	}

	if err = json.Unmarshal(buf, &reqBody); err != nil {
		rspBody["Err"] = err.Error()
		return
	}

	name := reqBody["Name"].(string)
	opts := reqBody["Opts"].(map[string]string)
	if size, err = strconv.ParseFloat(opts["Size"], 64); err != nil {
		rspBody["Err"] = err.Error()
	}
	v := volume.NewVolume(name, size)
	if v != nil {
		p.VolumeList[v.Name] = v
		go v.GetBackendStore(p.StoreServIP, p.StoreServPort, p.InitiatorName)
	}
	rspBody["Err"] = ""
}

func (p *Plugin) ListVolumes(rsp http.ResponseWriter, req *http.Request) {

}

func (p *Plugin) MountVolume(rsp http.ResponseWriter, req *http.Request) {

}

func (p *Plugin) RemoveVolume(rsp http.ResponseWriter, req *http.Request) {
	/*
		target := c.BlockList[strings.TrimSpace(dev)]

		if err := logoutTarget(target); err != nil {
			log.Fatal(err)
		}

		if err := removeBlock(target, c.StoreServIP, c.StoreServPort); err != nil {
			log.Fatal(err)
		}

		delete(c.BlockList, strings.TrimSpace(dev))
		p.ToDisk()
		fmt.Printf("removed block: %s\n", dev)
	*/
}

func (p *Plugin) VolumePath(rsp http.ResponseWriter, req *http.Request) {

}

func (p *Plugin) GetVolume(rsp http.ResponseWriter, req *http.Request) {

}
