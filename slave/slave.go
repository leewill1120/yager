package slave

import (
	"log"
	"net/http"

	"leewill1120/yager/drivers/lvm"
	"leewill1120/yager/drivers/rtslib"
)

type Slave struct {
	VG      *lvm.VolumeGroup
	RtsConf *rtslib.Config
}

func NewSlave() *Slave {
	s := &Slave{
		VG:      lvm.NewVG(""),
		RtsConf: rtslib.NewConfig(),
	}
	if s.RtsConf == nil || s.VG == nil {
		return nil
	} else {
		return s
	}
}

func (s *Slave) test(ResponseWriter http.ResponseWriter, Request *http.Request) {
	text := "This is a test page"
	ResponseWriter.Write([]byte(text))
}

func (s *Slave) Run() {
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/test", s.test)
	apiMux.HandleFunc("/block/create", s.CreateBlock)
	apiMux.HandleFunc("/block/delete", s.DeleteBlock)

	apiServer := &http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: apiMux,
	}

	log.Println("Slave listening on :8080")
	apiServer.ListenAndServe()
}
