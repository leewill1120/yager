package worker

import (
	"encoding/json"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"
)

//target represents block
func (s *Worker) ListBlock(rsp http.ResponseWriter, req *http.Request) {
	var (
		rspBody map[string]interface{} = make(map[string]interface{})
	)
	rspBody["result"] = "fail"

	if "slave" == s.WorkMode {
		cliIP := strings.Split(req.RemoteAddr, ":")[0]
		if !s.checkClientIP(cliIP) {
			rsp.Write([]byte("client ip check fail."))
			return
		}
	}

	defer func() {
		recover()
		if sendbuf, err := json.Marshal(rspBody); err != nil {
			log.WithFields(log.Fields{
				"reason": err,
				"data":   rspBody,
			}).Error("json.Marshal failed.")
		} else {
			rsp.Write(sendbuf)
		}
	}()

	targetList := make([]interface{}, 0)
	for _, t := range s.RtsConf.Targets {
		tt := make(map[string]interface{})
		tt["target"] = t.Wwn
		tt["host"] = strings.Split(req.Host, ":")[0]
		tt["port"] = t.Tpgs[0].Portals[0].Port
		targetList = append(targetList, tt)
	}
	rspBody["result"] = "success"
	rspBody["targetList"] = targetList
}
