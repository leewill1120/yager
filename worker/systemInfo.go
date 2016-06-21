package worker

import (
	"encoding/json"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"
)

//target represents block
func (s *Worker) SystemInfo(rsp http.ResponseWriter, req *http.Request) {
	var (
		rspBody map[string]interface{} = make(map[string]interface{})
	)
	rspBody["status"] = "stopped"

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

	rspBody["status"] = "running"
	rspBody["total"] = s.VG.Size
	rspBody["free"] = s.VG.Free
}
