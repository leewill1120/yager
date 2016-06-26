package plugin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"leewill1120/yager/drivers/iscsiadm"
	"leewill1120/yager/drivers/nfs"
	"leewill1120/yager/drivers/volume"

	log "github.com/Sirupsen/logrus"
)

var (
	defaultVolumeSize float64           = 1024 * 10
	supportType       map[string]string = map[string]string{"iscsi": "", "nfs": "", "cifs": ""}
)

func (p *Plugin) CreateVolume(rsp http.ResponseWriter, req *http.Request) {
	var (
		err        error
		buf        []byte                 = make([]byte, 1024)
		reqBody    map[string]interface{} = make(map[string]interface{})
		rspBody    map[string]interface{} = make(map[string]interface{})
		size       float64
		ok         bool
		vol        *volume.Volume
		volumeType string
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
	if _, ok = p.VolumeList[name]; ok {
		rspBody["Err"] = fmt.Sprintf("volume(%s) already exists.", name)
		return
	}

	if opts_interface, exists := reqBody["Opts"]; !exists {
		size = defaultVolumeSize
	} else {
		if nil == opts_interface {
			size = defaultVolumeSize
		} else {
			//get options
			opts := opts_interface.(map[string]interface{})
			if _, ok = opts["size"]; !ok {
				size = defaultVolumeSize
			} else {
				if size, err = strconv.ParseFloat(opts["size"].(string), 64); err != nil {
					rspBody["Err"] = err.Error()
					return
				}
			}

			if _, ok = opts["type"]; ok {
				volumeType = strings.TrimSpace(opts["type"].(string))
				if _, ok = supportType[volumeType]; !ok {
					rspBody["Err"] = fmt.Errorf("not supported type(%s)", volumeType)
					return
				}
			}

		}
	}

	vol, err = p.requestVolume(name, volumeType, size)
	if err != nil {
		rspBody["Err"] = err.Error()
		return
	}

	p.VolumeList[name] = vol
	p.ToDisk()
	rspBody["Err"] = ""
}

func (p *Plugin) releaseVolume(v *volume.Volume) error {
	var (
		err     error
		buf     []byte = make([]byte, 1024)
		method  reflect.Value
		reqBody map[string]interface{} = make(map[string]interface{})
		rspBody map[string]interface{} = make(map[string]interface{})
		url     string                 = "http://" + p.StoreMServIP + ":" + strconv.Itoa(p.StoreMServPort) + "/block/delete"
		rsp     *http.Response
	)

	method = reflect.ValueOf(*v).MethodByName("Attribute")
	if !method.IsValid() {
		return fmt.Errorf("method(Attribute) invalid.")
	}
	fn := method.Interface().(func() map[string]interface{})
	attributes := fn()

	reqBody["target"] = attributes["Target"]
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
		return fmt.Errorf("releaseVolume failed.reason:%s\n", rspBody["detail"])
	}
	return nil
}

func (p *Plugin) requestVolume(volumeName, volumeType string, size float64) (*volume.Volume, error) {
	var (
		err     error
		buf     []byte                 = make([]byte, 1024)
		reqBody map[string]interface{} = make(map[string]interface{})
		rspBody map[string]interface{} = make(map[string]interface{})
		url     string                 = "http://" + p.StoreMServIP + ":" + strconv.Itoa(p.StoreMServPort) + "/block/create"
		rsp     *http.Response
		vol     volume.Volume
	)

	reqBody["type"] = volumeType
	if "iscsi" == volumeType || "" == volumeType {
		reqBody["initiatorName"] = p.InitiatorName
		reqBody["size"] = size
	}

	if buf, err = json.Marshal(reqBody); err != nil {
		return nil, err
	}
	if rsp, err = http.Post(url, "application/json", bytes.NewBuffer(buf)); err != nil {
		return nil, err
	}
	if 4 == rsp.StatusCode/100 || 5 == rsp.StatusCode/100 {
		return nil, fmt.Errorf("server return %d.\n", rsp.StatusCode)
	}
	if buf, err = ioutil.ReadAll(rsp.Body); err != nil {
		return nil, err
	}

	if err = json.Unmarshal(buf, &rspBody); err != nil {
		log.WithFields(log.Fields{
			"data": string(buf),
		}).Debug(err)
		return nil, err
	}

	if "success" != rspBody["result"].(string) {
		return nil, fmt.Errorf("requestVolume failed.reason:%s\n", rspBody["detail"])
	}
	switch rspBody["type"].(string) {
	case "iscsi":
		if vol, err = iscsiadm.NewVolume(volumeName, rspBody); nil != err {
			log.WithFields(log.Fields{
				"reason": err,
			}).Debug("request volume faild.")
			return nil, err
		}
	case "nfs":
		if vol, err = nfs.NewVolume(volumeName, rspBody); nil != err {
			log.WithFields(log.Fields{
				"reason": err,
			}).Debug("request volume faild.")
			return nil, err
		}
	case "cifs":
		/*
			if vol, err = cifs.NewVolume(volumeName, rspBody); nil != err {
				log.WithFields(log.Fields{
					"reason": err,
				}).Debug("request volume faild.")
				return nil, err
			}
		*/
	default:
		return nil, fmt.Errorf("volume type(%s) not support", rspBody["type"].(string))
	}
	return &vol, nil
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
		method := reflect.ValueOf(*vl).MethodByName("Attribute")
		if !method.IsValid() {
			log.Fatal("method(Attribute) invalid.")
		}
		fn := method.Interface().(func() map[string]interface{})
		attributes := fn()

		vm := make(map[string]string)
		vm["Name"] = attributes["Name"].(string)
		vm["MountPoint"] = attributes["MountPoint"].(string)
		vm["Type"] = attributes["Type"].(string)
		volumeList = append(volumeList, vm)
	}

	rspBody["Volumes"] = volumeList
	rspBody["Err"] = ""
}

func (p *Plugin) MountVolume(rsp http.ResponseWriter, req *http.Request) {
	var (
		err     error
		buf     []byte = make([]byte, 1024)
		method  reflect.Value
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

	vl := p.VolumeList[name]

	method = reflect.ValueOf(*vl).MethodByName("Mount")
	if !method.IsValid() {
		rspBody["Err"] = "method(Mount) invalid."
		return
	}
	fnMount := method.Interface().(func() error)
	if err = fnMount(); err != nil {
		rspBody["Err"] = err.Error()
		return
	}

	method = reflect.ValueOf(*vl).MethodByName("Attribute")
	if !method.IsValid() {
		log.Fatal("method(Attribute) invalid.")
	}
	fnAttribute := method.Interface().(func() map[string]interface{})
	attributes := fnAttribute()

	rspBody["MountPoint"] = attributes["MountPoint"].(string)
	rspBody["Err"] = ""
}

func (p *Plugin) UmountVolume(rsp http.ResponseWriter, req *http.Request) {
	var (
		err     error
		buf     []byte = make([]byte, 1024)
		ok      bool
		reqBody map[string]interface{} = make(map[string]interface{})
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

	vl := p.VolumeList[name]
	method := reflect.ValueOf(*vl).MethodByName("Umount")
	if !method.IsValid() {
		rspBody["Err"] = "method(Umount) invalid."
		return
	}
	fn := method.Interface().(func() error)

	if err = fn(); err != nil {
		rspBody["Err"] = err.Error()
		return
	}
	rspBody["Err"] = ""
}

func (p *Plugin) RemoveVolume(rsp http.ResponseWriter, req *http.Request) {
	var (
		err     error
		buf     []byte = make([]byte, 1024)
		method  reflect.Value
		reqBody map[string]interface{} = make(map[string]interface{})
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

	if buf, err = ioutil.ReadAll(req.Body); err != nil {
		rspBody["Err"] = err.Error()
		return
	}

	if err = json.Unmarshal(buf, &reqBody); err != nil {
		rspBody["Err"] = err.Error()
		return
	}

	name := strings.TrimSpace(reqBody["Name"].(string))
	vl := p.VolumeList[name]
	method = reflect.ValueOf(*vl).MethodByName("Attribute")
	if !method.IsValid() {
		rspBody["Err"] = "method(Attribute) invalid."
		return
	}
	fn := method.Interface().(func() map[string]interface{})
	attributes := fn()

	if volume.MOUNTED == attributes["Status"].(int) {
		rspBody["Err"] = "volume is still mounted, couldn't remove."
		return
	}

	if err = p.releaseVolume(vl); err != nil {
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
	vl := p.VolumeList[name]
	method := reflect.ValueOf(*vl).MethodByName("Attribute")
	if !method.IsValid() {
		log.Fatal("method(Attribute) invalid.")
	}
	fn := method.Interface().(func() map[string]interface{})
	attributes := fn()

	rspBody["MountPoint"] = attributes["MountPoint"].(string)
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
	vl := p.VolumeList[name]
	method := reflect.ValueOf(*vl).MethodByName("Attribute")
	if !method.IsValid() {
		log.Fatal("method(Attribute) invalid.")
	}
	fn := method.Interface().(func() map[string]interface{})
	attributes := fn()

	vlm := make(map[string]string)
	vlm["Name"] = attributes["Name"].(string)
	vlm["MountPoint"] = attributes["MountPoint"].(string)
	rspBody["Volume"] = vlm
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
