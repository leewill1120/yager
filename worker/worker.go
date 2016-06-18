package worker

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"leewill1120/yager/drivers/lvm"
	"leewill1120/yager/drivers/rtslib"
)

type Worker struct {
	VG           *lvm.VolumeGroup
	RtsConf      *rtslib.Config
	ListenPort   int
	MasterIP     string
	MasterPort   int
	WorkMode     string
	RegisterCode string
}

func NewWorker(workmode, masterip string, masterport, listenport int, registercode string) *Worker {
	s := &Worker{
		VG:           lvm.NewVG(""),
		RtsConf:      rtslib.NewConfig(),
		ListenPort:   listenport,
		MasterIP:     masterip,
		MasterPort:   masterport,
		WorkMode:     workmode,
		RegisterCode: registercode,
	}
	if s.RtsConf == nil || s.VG == nil {
		return nil
	} else {
		return s
	}
}

func (s *Worker) test(ResponseWriter http.ResponseWriter, Request *http.Request) {
	text := "This is a test page"
	ResponseWriter.Write([]byte(text))
}

func (s *Worker) checkClientIP(clientIP string) bool {
	if strings.TrimSpace(s.MasterIP) == strings.TrimSpace(clientIP) {
		return true
	} else {
		return false
	}
}

func (s *Worker) Register() {
	for {
		msgBody := make(map[string]interface{})
		msgBody["registerCode"] = s.RegisterCode
		msgBody["port"] = s.ListenPort
		var (
			err error
			buf []byte
		)
		buf, err = json.Marshal(msgBody)
		if err != nil {
			log.Fatal(err)
		}

		body := bytes.NewBuffer(buf)
		if _, err := http.Post("http://"+s.MasterIP+":"+strconv.Itoa(s.MasterPort), "application/json", body); err != nil {
			log.Printf("register failed, reason:%s\n", err)
		}
		time.Sleep(time.Second * 10)
	}
}

func (s *Worker) Run(c chan error) {
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/test", s.test)
	apiMux.HandleFunc("/block/create", s.CreateBlock)
	apiMux.HandleFunc("/block/delete", s.DeleteBlock)
	apiMux.HandleFunc("/block/list", s.ListBlock)
	apiMux.HandleFunc("/system/info", s.SystemInfo)

	apiServer := &http.Server{
		Addr:    "0.0.0.0:" + strconv.Itoa(s.ListenPort),
		Handler: apiMux,
	}

	go func() {
		log.Println("Worker listening on " + apiServer.Addr)
		c <- apiServer.ListenAndServe()
	}()

	if "slave" == s.WorkMode {
		go s.Register()
	}
}
