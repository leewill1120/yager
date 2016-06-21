package plugin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"leewill1120/yager/plugin/volume"

	log "github.com/Sirupsen/logrus"
)

var (
	defaultVolumeSize float64 = 1024 * 10
)

func (p *Plugin) CreateVolume(rsp http.ResponseWriter, req *http.Request) {
	var (
		err     error
		buf     []byte                 = make([]byte, 1024)
		reqBody map[string]interface{} = make(map[string]interface{})
		rspBody map[string]string      = make(map[string]string)
		size    float64
		ok      bool
	)

	defer func() {
		if buf, err = json.Marshal(rspBody); err != nil {
			log.WithFields(log.Fields{
				"reason": err,
				"data":   rspBody,
			}).Error("json.Marshal failed.")
			return
		}
		if _, err = rsp.Write(buf); err != nil {
			log.Error(err)
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
	if opts_interface, exists := reqBody["Opts"]; !exists {
		size = defaultVolumeSize
	} else {
		if nil == opts_interface {
			size = defaultVolumeSize
		} else {
			opts := opts_interface.(map[string]interface{})
			if _, ok = opts["Size"]; !ok {
				size = defaultVolumeSize
			} else {
				if size, err = strconv.ParseFloat(opts["Size"].(string), 64); err != nil {
					rspBody["Err"] = err.Error()
					return
				}
			}
		}
	}

	v := volume.NewVolume(name, size)
	if v == nil {
		rspBody["Err"] = "error to get new volume."
		return
	}
	if err = v.GetBackendStore(p.StoreServIP, p.StoreServPort, p.InitiatorName); err != nil {
		rspBody["Err"] = err.Error()
		return
	}
	if err = v.LoginAndGetNewDev(); err != nil {
		rspBody["Err"] = err.Error()
		return
	}

	if err = v.FormatAndMount(); err != nil {
		rspBody["Err"] = err.Error()
		return
	}
	v.Status = volume.OK
	p.VolumeList[v.Name] = v

	p.ToDisk()
	rspBody["Err"] = ""
}

func (p *Plugin) ListVolumes(rsp http.ResponseWriter, req *http.Request) {
	var (
		err     error
		buf     []byte                 = make([]byte, 1024)
		rspBody map[string]interface{} = make(map[string]interface{})
	)

	defer func() {
		if buf, err = json.Marshal(rspBody); err != nil {
			log.WithFields(log.Fields{
				"reason": err,
				"data":   rspBody,
			}).Error("json.Marshal failed.")
			return
		}
		if _, err = rsp.Write(buf); err != nil {
			log.Error(err)
			return
		}
	}()

	volumeList := make([]interface{}, 0)
	for _, vl := range p.VolumeList {
		vm := make(map[string]string)
		vm["Name"] = vl.Name
		vm["MoMountpoint"] = vl.MountPoint
		volumeList = append(volumeList, vm)
	}

	rspBody["Volumes"] = volumeList
	rspBody["Err"] = ""
}

func (p *Plugin) MountVolume(rsp http.ResponseWriter, req *http.Request) {
	var (
		err     error
		buf     []byte                 = make([]byte, 1024)
		reqBody map[string]interface{} = make(map[string]interface{})
		rspBody map[string]interface{} = make(map[string]interface{})
		ok      bool
	)

	defer func() {
		if buf, err = json.Marshal(rspBody); err != nil {
			log.WithFields(log.Fields{
				"reason": err,
				"data":   rspBody,
			}).Error("json.Marshal failed.")
			return
		}
		if _, err = rsp.Write(buf); err != nil {
			log.Error(err)
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

	name := strings.TrimSpace(reqBody["Name"].(string))
	if _, ok = p.VolumeList[name]; !ok {
		rspBody["Err"] = fmt.Sprintf("volume(%s) not found.", name)
		return
	}
	rspBody["Mountpoint"] = p.VolumeList[name].MountPoint
	rspBody["Err"] = ""
}

func (p *Plugin) UnmountVolume(rsp http.ResponseWriter, req *http.Request) {
	var (
		err     error
		buf     []byte                 = make([]byte, 1024)
		rspBody map[string]interface{} = make(map[string]interface{})
	)

	defer func() {
		if buf, err = json.Marshal(rspBody); err != nil {
			log.WithFields(log.Fields{
				"reason": err,
				"data":   rspBody,
			}).Error("json.Marshal failed.")
			return
		}
		if _, err = rsp.Write(buf); err != nil {
			log.Error(err)
			return
		}
	}()
	rspBody["Err"] = ""
}

func (p *Plugin) RemoveVolume(rsp http.ResponseWriter, req *http.Request) {
	var (
		err     error
		buf     []byte                 = make([]byte, 1024)
		reqBody map[string]interface{} = make(map[string]interface{})
		rspBody map[string]string      = make(map[string]string)
	)

	defer func() {
		if buf, err = json.Marshal(rspBody); err != nil {
			log.WithFields(log.Fields{
				"reason": err,
				"data":   rspBody,
			}).Error("json.Marshal failed.")
			return
		}
		if _, err = rsp.Write(buf); err != nil {
			log.Error(err)
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

	name := strings.TrimSpace(reqBody["Name"].(string))
	if err = p.VolumeList[name].Umount(); err != nil {
		rspBody["Err"] = err.Error()
		return
	}
	if err = p.VolumeList[name].LogoutTarget(); err != nil {
		rspBody["Err"] = err.Error()
		return
	}
	if err = p.VolumeList[name].ReleaseBackendStore(p.StoreServIP, p.StoreServPort); err != nil {
		rspBody["Err"] = err.Error()
		return
	}

	delete(p.VolumeList, name)
	p.ToDisk()
	rspBody["Err"] = ""
}

func (p *Plugin) VolumePath(rsp http.ResponseWriter, req *http.Request) {
	var (
		err     error
		buf     []byte                 = make([]byte, 1024)
		reqBody map[string]interface{} = make(map[string]interface{})
		rspBody map[string]interface{} = make(map[string]interface{})
		ok      bool
	)

	defer func() {
		if buf, err = json.Marshal(rspBody); err != nil {
			log.WithFields(log.Fields{
				"reason": err,
				"data":   rspBody,
			}).Error("json.Marshal failed.")
			return
		}
		if _, err = rsp.Write(buf); err != nil {
			log.Error(err)
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

	name := strings.TrimSpace(reqBody["Name"].(string))
	if _, ok = p.VolumeList[name]; !ok {
		rspBody["Err"] = fmt.Sprintf("volume(%s) not found.", name)
		return
	}
	rspBody["Mountpoint"] = p.VolumeList[name].MountPoint
	rspBody["Err"] = ""
}

func (p *Plugin) GetVolume(rsp http.ResponseWriter, req *http.Request) {
	var (
		err     error
		buf     []byte                 = make([]byte, 1024)
		reqBody map[string]interface{} = make(map[string]interface{})
		rspBody map[string]interface{} = make(map[string]interface{})
		ok      bool
	)

	defer func() {
		if buf, err = json.Marshal(rspBody); err != nil {
			log.WithFields(log.Fields{
				"reason": err,
				"data":   rspBody,
			}).Error("json.Marshal failed.")
			return
		}
		if _, err = rsp.Write(buf); err != nil {
			log.Error(err)
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

	name := strings.TrimSpace(reqBody["Name"].(string))
	if _, ok = p.VolumeList[name]; !ok {
		rspBody["Err"] = fmt.Sprintf("volume(%s) not found.", name)
		return
	}
	vl := make(map[string]string)
	vl["Name"] = name
	vl["Mountpoint"] = p.VolumeList[name].MountPoint
	rspBody["Volume"] = vl
	rspBody["Err"] = ""
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
	url := "http://" + ip + ":" + port + "/block/delete"
	if rsp, err := http.Post(url, "application/json", body); err != nil {
		log.WithFields(log.Fields{
			"host":   ip + ":" + port,
			"url":    url,
			"reason": err,
		}).Warn("delete block failed.")
		return err
	} else {
		if 4 == rsp.StatusCode/100 || 5 == rsp.StatusCode {
			log.WithFields(log.Fields{
				"host":       ip + ":" + port,
				"url":        url,
				"StatusCode": rsp.StatusCode,
			}).Warn("delete block failed.")
			return fmt.Errorf("Worker return %d", rsp.StatusCode)
		}

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
