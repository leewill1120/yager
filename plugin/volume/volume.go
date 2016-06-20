package volume

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

const (
	INIT = iota
	OK
	BAD
)

type Volume struct {
	Name          string
	MountPoint    string
	Target        string
	StoreServIP   string
	StoreServPort int
	UserID        string
	Password      string
	Size          float64
	Status        int
}

func NewVolume(volumeName string, size float64) *Volume {
	return &Volume{
		Name:   volumeName,
		Size:   size,
		Status: INIT,
	}
}

func (v *Volume) GetBackendStore(storeMServIP string, storeMServPort int, initiatorName string) error {
	var (
		err     error
		buf     []byte                 = make([]byte, 1024)
		reqBody map[string]interface{} = make(map[string]interface{})
		rspBody map[string]interface{} = make(map[string]interface{})
		url     string                 = "http://" + storeMServIP + ":" + strconv.Itoa(storeMServPort) + "/block/create"
		rsp     *http.Response
	)

	reqBody["InitiatorName"] = initiatorName
	reqBody["Size"] = v.Size
	if buf, err = json.Marshal(reqBody); err != nil {
		return err
	}
	if rsp, err = http.Post(url, "application/json", bytes.NewBuffer(buf)); err != nil {
		return err
	}
	if 4 == rsp.StatusCode/100 || 5 == rsp.StatusCode/100 {
		return fmt.Errorf("server return %d.", rsp.StatusCode)
	}
	if buf, err = ioutil.ReadAll(rsp.Body); err != nil {
		return err
	}

	if err = json.Unmarshal(buf, &rspBody); err != nil {
		return err
	}

	if "success" != rspBody["result"].(string) {
		return fmt.Errorf("GetBackendStore failed.reason:%s ", rspBody["detail"])
	}

	v.StoreServIP = rspBody["host"].(string)
	v.StoreServPort = int(rspBody["port"].(float64))
	v.UserID = rspBody["userid"].(string)
	v.Password = rspBody["password"].(string)
	v.Target = rspBody["target"].(string)

	return nil
}
