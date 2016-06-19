package manager

import (
	"encoding/json"
	"log"
	"net/http"
)

func (m *Manager) DeleteBlock(rsp http.ResponseWriter, req *http.Request) {
	var (
		err     error
		buf     []byte
		rspBody map[string]interface{} = make(map[string]interface{})
	)

	rspBody["result"] = "fail"
	defer func() {
		if buf, err = json.Marshal(rspBody); err != nil {
			log.Println(err)
		} else {
			rsp.Write(buf)
		}
	}()

	list := make([]map[string]interface{}, 0)
	for k, v := range m.WorkerList {
		w := make(map[string]interface{})
		w["ip"] = k
		w["port"] = v.Port
		list = append(list, w)
	}
	rspBody["workerList"] = list
	rspBody["result"] = "success"
}
