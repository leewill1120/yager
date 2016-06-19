package manager

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"leewill1120/yager/manager/worker"
)

var (
	aliveTimeout time.Duration = time.Second * 15
)

func (m *Manager) WorkerRegister(rsp http.ResponseWriter, req *http.Request) {
	var (
		err     error
		buf     []byte
		reqBody map[string]interface{} = make(map[string]interface{})
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

	buf, err = ioutil.ReadAll(req.Body)
	if err != nil {
		rspBody["detail"] = "invalid argument."
		log.Println(err)
		return
	}
	if err = json.Unmarshal(buf, &reqBody); err != nil {
		rspBody["detail"] = "invalid argument."
		log.Println(err)
		return
	}

	if rc, ok := reqBody["registerCode"]; !ok {
		rspBody["detail"] = "argument registerCode doesn't exist."
		return
	} else {
		if m.RegisterCode != strings.TrimSpace(rc.(string)) {
			rspBody["detail"] = "registerCode doesn't match."
			return
		}
	}

	if _, ok := reqBody["port"]; !ok {
		rspBody["detail"] = "argument port doesn't exist."
		return
	}

	workerIP := strings.Split(req.RemoteAddr, ":")[0]
	workerPort := int(reqBody["port"].(float64))
	if _, ok := m.WorkerList[workerIP]; ok {
		m.WorkerList[workerIP].Port = workerPort
		m.WorkerList[workerIP].Timer.Reset(aliveTimeout)
	} else {
		m.WorkerList[workerIP] = &worker.Worker{
			IP:    workerIP,
			Port:  workerPort,
			Timer: time.NewTimer(aliveTimeout),
		}
		log.Printf("New worker(%s:%d) joined", workerIP, workerPort)
		go func() {
			<-m.WorkerList[workerIP].Timer.C
			delete(m.WorkerList, workerIP)
			log.Printf("worker(%s:%d) leave.\n", workerIP, workerPort)
		}()
	}
	rspBody["result"] = "success"
}
