package worker

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func (s *Worker) DeleteBlock(ResponseWriter http.ResponseWriter, Request *http.Request) {
	var (
		err     error
		target  string
		reqBody map[string]interface{} = make(map[string]interface{})
		rspBody map[string]interface{} = make(map[string]interface{})
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
			log.Println(err)
		} else {
			ResponseWriter.Write(sendbuf)
		}
	}()

	reqMsg := make([]byte, 1024)
	if reqMsg, err = ioutil.ReadAll(Request.Body); err != nil {
		log.Println(err)
		rspBody["result"] = "fail"
		rspBody["detail"] = "invalid argument."
		return
	}

	if err = json.Unmarshal(reqMsg, &reqBody); err != nil {
		log.Println(err)
		rspBody["result"] = "fail"
		rspBody["detail"] = "invalid argument."
		return
	}

	if target_i, exist_i := reqBody["target"]; !exist_i {
		rspBody["result"] = "fail"
		rspBody["detail"] = "argument target not exist."
		return
	} else {
		target = target_i.(string)
	}

	t := s.RtsConf.GetTarget(target)
	if nil == t {
		rspBody["result"] = "fail"
		rspBody["detail"] = "target not exist."
		return
	}
	soPath := t.Tpgs[0].Luns[0].Storage_object
	soName := strings.Split(soPath, "/")[len(strings.Split(soPath, "/"))-1]
	so := s.RtsConf.GetStore(soName)
	if nil == so {
		rspBody["result"] = "fail"
		rspBody["detail"] = "Storage_object not exist."
		return
	}
	lvPath := so.Dev
	lvName := strings.Split(lvPath, "/")[len(strings.Split(lvPath, "/"))-1]

	if err = s.RtsConf.RemoveTarget(target); err != nil {
		log.Println(err)
		rspBody["result"] = "fail"
		rspBody["detail"] = err.Error()
		return
	}

	if err = s.RtsConf.RemoveStore(soName); err != nil {
		log.Println(err)
		rspBody["result"] = "fail"
		rspBody["detail"] = err.Error()
		return
	}

	//notify to apply setting
	c := make(chan error)
	s.ApplyChan <- c
	err = <-c
	if err != nil {
		log.Println(err)
		rspBody["result"] = "fail"
		rspBody["detail"] = err.Error()
		return
	}

	if err = s.VG.RemoveLV(lvName); err != nil {
		log.Println(err)
		rspBody["result"] = "fail"
		rspBody["detail"] = err.Error()
		return
	}

	rspBody["result"] = "success"
}
