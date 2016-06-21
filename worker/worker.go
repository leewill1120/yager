package worker

import (
	"leewill1120/mux"
	"leewill1120/yager/drivers/lvm"
	"leewill1120/yager/drivers/rtslib"
	"net/http"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
)

type Worker struct {
	VG           *lvm.VolumeGroup
	RtsConf      *rtslib.Config
	ApplyChan    chan chan error
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
		ApplyChan:    make(chan chan error),
	}
	if s.RtsConf == nil || s.VG == nil {
		return nil
	} else {
		return s
	}
}

func (s *Worker) checkClientIP(clientIP string) bool {
	if strings.TrimSpace(s.MasterIP) == strings.TrimSpace(clientIP) {
		return true
	} else {
		return false
	}
}

func (s *Worker) Run(c chan error) {
	go func() {
		for {
			c := <-s.ApplyChan

			if err := s.RtsConf.ToDisk(""); err != nil {
				c <- err
				return
			}

			if err := s.RtsConf.Restore(""); err != nil {
				c <- err
				return
			}
			c <- nil
		}
	}()

	router := mux.NewRouter()
	router.HandleFunc("/block/create", s.CreateBlock).Methods("POST")
	router.HandleFunc("/block/delete", s.DeleteBlock).Methods("POST")
	router.HandleFunc("/block/list", s.ListBlock).Methods("GET")
	router.HandleFunc("/system/info", s.SystemInfo).Methods("GET")

	apiServer := &http.Server{
		Addr:    "0.0.0.0:" + strconv.Itoa(s.ListenPort),
		Handler: router,
	}

	go func() {
		log.Infof("Worker listening on %s.", apiServer.Addr)
		c <- apiServer.ListenAndServe()
	}()

	if "slave" == s.WorkMode {
		go s.Register()
	}
}
