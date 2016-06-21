package worker

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/pborman/uuid"
)

func (s *Worker) CreateBlock(ResponseWriter http.ResponseWriter, Request *http.Request) {
	var (
		err                       error
		initiatorName, lv, target string
		size                      float64
		exist, ok                 bool
		size_interface            interface{}
		buf                       []byte                 = make([]byte, 1024)
		msgBody                   map[string]interface{} = make(map[string]interface{})
		rspBody                   map[string]interface{} = make(map[string]interface{})
		username                  string                 = uuid.New()[24:]
		password                  string                 = uuid.New()[24:]
	)
	if "slave" == s.WorkMode {
		cliIP := strings.Split(Request.RemoteAddr, ":")[0]
		if !s.checkClientIP(cliIP) {
			ResponseWriter.Write([]byte("client ip check fail."))
			return
		}
	}

	defer func() {
		if sendbuf, err := json.Marshal(rspBody); err != nil {
			log.WithFields(log.Fields{
				"reason": err,
				"data":   rspBody,
			}).Error("json.Marshal failed.")
		} else {
			ResponseWriter.Write(sendbuf)
		}
	}()

	buf, err = ioutil.ReadAll(Request.Body)
	if err != nil {
		rspBody["result"] = "fail"
		rspBody["detail"] = "invalid argument."
		log.Error(err)
		return
	} else {
		if e := json.Unmarshal(buf, &msgBody); e != nil {
			rspBody["result"] = "fail"
			rspBody["detail"] = "invalid argument."
			log.WithFields(log.Fields{
				"reason": err,
				"data":   string(buf),
			}).Error("json.Unmarshal failed.")
			return
		}
	}

	if initiatorName_i, exist_i := msgBody["initiatorName"]; !exist_i {
		rspBody["result"] = "fail"
		rspBody["detail"] = "argument initiatorName not exist."
		return
	} else {
		initiatorName = initiatorName_i.(string)
	}

	if size_interface, exist = msgBody["size"]; !exist {
		rspBody["result"] = "fail"
		rspBody["detail"] = "argument size not exist."
		return
	} else {
		if size, ok = size_interface.(float64); !ok {
			rspBody["result"] = "fail"
			rspBody["detail"] = "error to parse size."
			return
		}
	}

	if lv, err = s.VG.CreateLV(size); err != nil {
		rspBody["result"] = "fail"
		rspBody["detail"] = err.Error()
		log.Error(err)
		return
	}

	blockName := "block." + lv
	if err = s.RtsConf.AddBlockStore("/dev/"+s.VG.Name+"/"+lv, blockName); err != nil {
		rspBody["result"] = "fail"
		rspBody["detail"] = err.Error()
		log.Error(err)
		return
	}

	storage_object := "/backstores/block/" + blockName
	if target, err = s.RtsConf.AddTarget(storage_object, initiatorName, username, password); err != nil {
		rspBody["result"] = "fail"
		rspBody["detail"] = err.Error()
		log.Error(err)
		return
	}

	//notify to apply setting
	c := make(chan error)
	s.ApplyChan <- c
	err = <-c
	if err != nil {
		rspBody["result"] = "fail"
		rspBody["detail"] = err.Error()
		log.Error(err)
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
