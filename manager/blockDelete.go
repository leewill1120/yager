package manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func (m *Manager) DeleteBlock(rsp http.ResponseWriter, req *http.Request) {
	var (
		err     error
		buf     []byte
		ok      bool
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

	if buf, err = ioutil.ReadAll(req.Body); err != nil {
		rspBody["detail"] = err.Error()
		return
	}

	if err = json.Unmarshal(buf, &reqBody); err != nil {
		rspBody["detail"] = err.Error()
		return
	}

	if _, ok = reqBody["target"]; !ok {
		rspBody["detail"] = "argument target doesn't exist."
		return
	}

	targetName := strings.TrimSpace(reqBody["target"].(string))
	if _, ok = m.TargetWorkerList[targetName]; !ok {
		rspBody["detail"] = fmt.Sprintf("couldn't find target(%s) info.", targetName)
		return
	}

	worker := m.WorkerList[m.TargetWorkerList[targetName]]
	if rsp2, err := http.Post("http://"+worker.IP+":"+strconv.Itoa(worker.Port)+"/block/delete", "application/json", bytes.NewBuffer(buf)); err == nil {
		if 4 == rsp2.StatusCode/100 || 5 == rsp2.StatusCode {
			rspBody["detail"] = fmt.Sprintln("Worker return %d", rsp2.StatusCode)
			return
		}
		if buf, err = ioutil.ReadAll(rsp2.Body); err != nil {
			rspBody["detail"] = err.Error()
			log.Println(err)
			return
		}

		if err = json.Unmarshal(buf, &rspBody); err != nil {
			rspBody["result"] = "fail"
			rspBody["detail"] = err.Error()
			log.Println(err, string(buf))
			return
		}

		if "success" == rspBody["result"].(string) {
			delete(m.TargetWorkerList, targetName)
		}

	} else {
		rspBody["detail"] = err.Error()
		return
	}
}
