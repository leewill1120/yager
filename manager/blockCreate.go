package manager

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"leewill1120/yager/manager/worker"
	"log"
	"net/http"
	"sort"
	"strconv"
)

func (m *Manager) CreateBlock(rsp http.ResponseWriter, req *http.Request) {
	var (
		err            error
		size           float64
		exist, ok      bool
		size_interface interface{}
		buf            []byte                 = make([]byte, 1024)
		reqBody        map[string]interface{} = make(map[string]interface{})
		rspBody        map[string]interface{} = make(map[string]interface{})
	)

	defer func() {
		if sendbuf, err := json.Marshal(rspBody); err != nil {
			log.Println(err)
		} else {
			rsp.Write(sendbuf)
		}
	}()

	buf, err = ioutil.ReadAll(req.Body)
	if err != nil {
		rspBody["result"] = "fail"
		rspBody["detail"] = "invalid argument."
		return
	} else {
		if e := json.Unmarshal(buf, &reqBody); e != nil {
			rspBody["result"] = "fail"
			rspBody["detail"] = "invalid argument."
			return
		}
	}

	if size_interface, exist = reqBody["Size"]; !exist {
		rspBody["result"] = "fail"
		rspBody["detail"] = "argument Size not exist."
		return
	}

	if size, ok = size_interface.(float64); !ok {
		rspBody["result"] = "fail"
		rspBody["detail"] = "error to parse Size."
		return
	}

	for _, w := range m.WorkerList {
		w.GetCapInfo()
	}

	//free bigger than request size and usage is lowest
	availableList := worker.WorkerList{}
	for _, w := range m.WorkerList {
		if size <= w.Free {
			availableList.List = append(availableList.List, w)
		}
	}

	sort.Sort(availableList)

	for _, w := range availableList.List {
		buf, _ = json.Marshal(reqBody)
		if rsp, err := http.Post("http://"+w.IP+":"+strconv.Itoa(w.Port)+"/block/create", "application/json", bytes.NewBuffer(buf)); err == nil {
			if (rsp.StatusCode/100 == 4) || (rsp.StatusCode/100 == 5) {
				log.Printf("server return %d.", rsp.StatusCode)
				continue
			}
			buf, err = ioutil.ReadAll(rsp.Body)
			if err != nil {
				log.Println(err)
				continue
			}
			err = json.Unmarshal(buf, &rspBody)
			if err != nil {
				log.Println(err, string(buf))
				continue
			}
			if "success" != rspBody["result"].(string) {
				log.Println(rspBody["detail"])
				continue
			}
			//success
			break
		} else {
			log.Println(err)
			continue
		}
	}
}
