package worker

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

//target represents block
func (s *Worker) SystemInfo(rsp http.ResponseWriter, req *http.Request) {
	var (
		rspBody map[string]interface{} = make(map[string]interface{})
	)
	rspBody["status"] = "stopped"

	if "slave" == s.WorkMode {
		cliIP := strings.Split(req.Host, ":")[0]
		if !s.checkClientIP(cliIP) {
			rsp.Write([]byte("404 not found"))
			return
		}
	}
	if "GET" != req.Method {
		rsp.Write([]byte("method not found"))
		return
	}

	defer func() {
		recover()
		if sendbuf, err := json.Marshal(rspBody); err != nil {
			log.Println(err)
		} else {
			rsp.Write(sendbuf)
		}
	}()

	rspBody["status"] = "running"
	rspBody["total"] = s.VG.Size
	rspBody["free"] = s.VG.Free
}
