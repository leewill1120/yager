package manager

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

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
			log.WithFields(log.Fields{
				"reason": err,
				"data":   rspBody,
			}).Error("json.Marshal failed.")
		} else {
			rsp.Write(buf)
		}
	}()

	buf, err = ioutil.ReadAll(req.Body)
	if err != nil {
		rspBody["detail"] = "invalid argument."
		log.Error(err)
		return
	}
	if err = json.Unmarshal(buf, &reqBody); err != nil {
		rspBody["detail"] = "invalid argument."
		log.WithFields(log.Fields{
			"reason": err,
			"data":   string(buf),
		}).Error("json.Unmarshal failed.")
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
		log.Infof("New worker(%s:%d) joined", workerIP, workerPort)
		go func() {
			<-m.WorkerList[workerIP].Timer.C
			delete(m.WorkerList, workerIP)
			for k, v := range m.TargetWorkerList {
				if v == workerIP {
					delete(m.TargetWorkerList, k)
				}
			}
			log.Infof("worker(%s:%d) leave.\n", workerIP, workerPort)
		}()

		//Get targets on this worker.
		rsp2Body := make(map[string]interface{})
		url := "http://" + m.WorkerList[workerIP].IP + ":" + strconv.Itoa(m.WorkerList[workerIP].Port) + "/block/list"
		if rsp2, err := http.Get(url); err != nil {
			log.WithFields(log.Fields{
				"url":    url,
				"reason": err,
				"worker": workerIP,
			}).Error("get targets on this worker failed.")
		} else {
			if 4 == rsp2.StatusCode/100 || 5 == rsp2.StatusCode {
				log.WithFields(log.Fields{
					"url":        url,
					"statusCode": rsp2.StatusCode,
					"worker":     workerIP,
				}).Error("get targets on this worker failed.")
				goto end
			}
			if buf, err = ioutil.ReadAll(rsp2.Body); err != nil {
				log.Error(err)
				goto end
			}

			if err = json.Unmarshal(buf, &rsp2Body); err != nil {
				log.WithFields(log.Fields{
					"reason": err,
					"data":   string(buf),
				}).Error("json.Unmarshal failed.")
				goto end
			}

			if "success" != rsp2Body["result"].(string) {
				log.WithFields(log.Fields{
					"url":    url,
					"reason": rsp2Body["detail"],
					"worker": workerIP,
				}).Error("get targets on this worker failed.")
				goto end
			}

			//success  to get targets
			for _, t := range rsp2Body["targetList"].([]interface{}) {
				tt := t.(map[string]interface{})
				m.TargetWorkerList[tt["target"].(string)] = tt["host"].(string)
			}
		}
	}
end:
	rspBody["result"] = "success"
}
