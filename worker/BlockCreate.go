package worker

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/pborman/uuid"
)

func (s *Worker) CreateBlock(ResponseWriter http.ResponseWriter, Request *http.Request) {
	var (
		err                       error
		initiatorName, lv, target string
		size                      float64
		exist, ok                 bool
		size_interface            interface{}
		msgBody                   []byte                 = make([]byte, 1024)
		argsMap                   map[string]interface{} = make(map[string]interface{})
		rspBody                   map[string]interface{} = make(map[string]interface{})
		username                  string                 = uuid.New()[24:]
		password                  string                 = uuid.New()[24:]
	)
	if "slave" == s.WorkMode {
		cliIP := strings.Split(Request.Host, ":")[0]
		if !s.checkClientIP(cliIP) {
			ResponseWriter.Write([]byte("404 not found"))
			return
		}
	}

	if "POST" != Request.Method {
		ResponseWriter.Write([]byte("method not found"))
		return
	}

	defer func() {
		if sendbuf, err := json.Marshal(rspBody); err != nil {
			log.Println(err)
		} else {
			ResponseWriter.Write(sendbuf)
		}
	}()

	msgBody, err = ioutil.ReadAll(Request.Body)
	if err != nil {
		rspBody["result"] = "fail"
		rspBody["detail"] = err.Error()
		return
	} else {
		if e := json.Unmarshal(msgBody, &argsMap); e != nil {
			rspBody["result"] = "fail"
			rspBody["detail"] = e.Error()
			return
		}
	}

	if initiatorName_i, exist_i := argsMap["InitiatorName"]; !exist_i {
		rspBody["result"] = "fail"
		rspBody["detail"] = "argument InitiatorName not exist."
		return
	} else {
		initiatorName = initiatorName_i.(string)
	}

	if size_interface, exist = argsMap["Size"]; !exist {
		rspBody["result"] = "fail"
		rspBody["detail"] = "argument Size not exist."
		return
	} else {
		if size, ok = size_interface.(float64); !ok {
			rspBody["result"] = "fail"
			rspBody["detail"] = "error to parse Size."
			return
		}
	}

	if lv, err = s.VG.CreateLV(size); err != nil {
		log.Println(err)
		rspBody["result"] = "fail"
		rspBody["detail"] = err.Error()
		return
	}

	blockName := "block." + lv
	if err = s.RtsConf.AddBlockStore("/dev/"+s.VG.Name+"/"+lv, blockName); err != nil {
		log.Println(err)
		rspBody["result"] = "fail"
		rspBody["detail"] = err.Error()
		return
	}

	storage_object := "/backstores/block/" + blockName
	if target, err = s.RtsConf.AddTarget(storage_object, initiatorName, username, password); err != nil {
		log.Println(err)
		rspBody["result"] = "fail"
		rspBody["detail"] = err.Error()
		return
	}

	if err = s.RtsConf.ToDisk(""); err != nil {
		log.Println(err)
		rspBody["result"] = "fail"
		rspBody["detail"] = err.Error()
		return
	}

	if err = s.RtsConf.Restore(""); err != nil {
		log.Println(err)
		rspBody["result"] = "fail"
		rspBody["detail"] = err.Error()
		return
	}

	rspBody["result"] = "success"
	rspBody["target"] = target
	rspBody["userid"] = username
	rspBody["password"] = password
	rspBody["host"] = strings.Split(Request.Host, ":")[0]
	rspBody["port"] = -1
	if stTarget := s.RtsConf.GetTarget(target); stTarget != nil {
		rspBody["port"] = stTarget.Tpgs[0].Portals[0].Port
	}
}
