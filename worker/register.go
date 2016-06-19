package worker

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

var (
	keepAliveInterval time.Duration = time.Second * 5
)

func (s *Worker) Register() {
	for {
		msgBody := make(map[string]interface{})
		msgBody["registerCode"] = s.RegisterCode
		msgBody["port"] = s.ListenPort
		rspBody := make(map[string]interface{})
		var (
			err error
			buf []byte
		)
		buf, err = json.Marshal(msgBody)
		if err != nil {
			log.Fatal(err)
		}

		body := bytes.NewBuffer(buf)
		if rsp, err := http.Post("http://"+s.MasterIP+":"+strconv.Itoa(s.MasterPort)+"/worker/register", "application/json", body); err != nil {
			log.Printf("register failed, reason:%s\n", err)
			time.Sleep(time.Second)
			continue
		} else {
			if (rsp.StatusCode/100 == 4) || (rsp.StatusCode/100 == 5) {
				log.Printf("server return %d.", rsp.StatusCode)
				time.Sleep(time.Second)
				continue
			}
			buf, err = ioutil.ReadAll(rsp.Body)
			if err != nil {
				log.Println(err)
			} else {
				if err = json.Unmarshal(buf, &rspBody); err != nil {
					log.Println(err)
				} else {
					if "success" != rspBody["result"].(string) {
						log.Println(rspBody["detail"].(string))
					}
				}
			}
		}
		time.Sleep(keepAliveInterval)
	}
}
