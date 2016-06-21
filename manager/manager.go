package manager

import (
	"leewill1120/mux"
	"leewill1120/yager/manager/worker"
	"net/http"
	"strconv"

	log "github.com/Sirupsen/logrus"
)

type Target struct {
	Name     string
	Worker   string // ip address of worker
	UserID   string
	Password string
}

type Manager struct {
	ListenPort       int
	RegisterCode     string
	ClientToken      string
	WorkerList       map[string]*worker.Worker //The key is ip address of worker
	TargetWorkerList map[string]string         //The key is target, value is ip address of worker
}

func NewManager(listenport int, registercode string) *Manager {
	m := &Manager{
		ListenPort:       listenport,
		RegisterCode:     registercode,
		WorkerList:       make(map[string]*worker.Worker),
		TargetWorkerList: make(map[string]string),
	}
	return m
}

func (m *Manager) Run(c chan interface{}) {
	router := mux.NewRouter()
	router.HandleFunc("/worker/register", m.WorkerRegister).Methods("POST")
	router.HandleFunc("/worker/list", m.GetWorkerList).Methods("GET")
	router.HandleFunc("/block/create", m.CreateBlock).Methods("POST")
	router.HandleFunc("/block/delete", m.DeleteBlock).Methods("POST")

	apiServer := &http.Server{
		Addr:    "0.0.0.0:" + strconv.Itoa(m.ListenPort),
		Handler: router,
	}
	go func() {
		log.Infof("manager listening on %s.", apiServer.Addr)
		c <- apiServer.ListenAndServe()
	}()
}
