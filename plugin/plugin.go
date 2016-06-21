package plugin

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"

	log "github.com/Sirupsen/logrus"

	"leewill1120/mux"
	"leewill1120/yager/plugin/volume"
)

var (
	defaultInitiatorNameFile string = "/etc/iscsi/initiatorname.iscsi"
	defaultUnixDomainSocket  string = "/run/docker/plugins/yager.sock"
)

type Block struct {
	IP       string
	Target   string
	Username string
	Password string
}

type Plugin struct {
	StoreServIP   string
	StoreServPort int
	VolumeList    map[string]*volume.Volume
	InitiatorName string
}

func NewPlugin(storeServIP string, storeServPort int) *Plugin {
	p := &Plugin{
		StoreServIP:   storeServIP,
		StoreServPort: storeServPort,
		VolumeList:    make(map[string]*volume.Volume),
	}

	if d, e := ioutil.ReadFile(defaultInitiatorNameFile); e != nil {
		log.Fatal(e)
	} else {
		InitiatorName := strings.Replace(string(d), "\n", "", -1)
		InitiatorName = strings.Replace(string(d), "InitiatorName=", "", -1)
		InitiatorName = strings.TrimSpace(InitiatorName)
		p.InitiatorName = InitiatorName
	}
	if err := p.FromDisk(); err != nil {
		log.Fatal(err)
	}
	return p
}

func (p *Plugin) Run() {
	p.createUnixSock(defaultUnixDomainSocket)

	router := mux.NewRouter()
	router.HandleFunc("/Plugin.Activate", p.Activate).Methods("POST")
	router.HandleFunc("/VolumeDriver.Create", p.CreateVolume).Methods("POST")
	router.HandleFunc("/VolumeDriver.Remove", p.RemoveVolume).Methods("POST")
	router.HandleFunc("/VolumeDriver.Mount", p.MountVolume).Methods("POST")
	router.HandleFunc("/VolumeDriver.Unmount", p.UnmountVolume).Methods("POST")
	router.HandleFunc("/VolumeDriver.Path", p.VolumePath).Methods("POST")
	router.HandleFunc("/VolumeDriver.Get", p.GetVolume).Methods("POST")
	router.HandleFunc("/VolumeDriver.List", p.ListVolumes).Methods("POST")

	addr := &net.UnixAddr{
		Name: defaultUnixDomainSocket,
		Net:  "unix",
	}
	l, e := net.ListenUnix("unix", addr)
	if e != nil {
		log.Fatal(e)
	}
	http.Serve(l, router)
}

func (p *Plugin) createUnixSock(unixsock string) error {
	if _, err := os.Stat(unixsock); err != nil {
		os.MkdirAll(path.Dir(unixsock), os.ModeDir)
	}
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	go func() {
		fmt.Printf("Received stop signal(%s)\n", <-c)
		os.Remove(unixsock)
		os.Exit(0)
	}()
	return nil
}

func (p *Plugin) Activate(rsp http.ResponseWriter, req *http.Request) {
	rspBody := make(map[string]interface{})
	implements := make([]string, 0)
	implements = append(implements, "VolumeDriver")
	rspBody["Implements"] = implements
	if buf, err := json.Marshal(rspBody); err != nil {
		log.WithFields(log.Fields{
			"reason": err,
			"data":   rspBody,
		}).Error("json.Marshal failed.")
		rsp.Write([]byte(err.Error()))
	} else {
		_, err = rsp.Write(buf)
		if err != nil {
			log.Error(err)
		}
	}
}

func (p *Plugin) FromDisk() error {
	home := os.Getenv("HOME")
	configPath := home + "/.yager.json"
	if _, err := os.Stat(configPath); err != nil {
		return p.ToDisk()
	}
	if d, e := ioutil.ReadFile(configPath); e != nil {
		return e
	} else {
		if e = json.Unmarshal(d, &p.VolumeList); e != nil {
			log.WithFields(log.Fields{
				"reason": e,
				"data":   string(d),
			}).Error("json.Unmarshal failed.")
			return e
		}
	}
	return nil
}

func (p *Plugin) ToDisk() error {
	home := os.Getenv("HOME")
	configPath := home + "/.yager.json"
	if b, err := json.Marshal(p.VolumeList); err != nil {
		log.WithFields(log.Fields{
			"reason": err,
			"data":   p.VolumeList,
		}).Error("json.Marshal failed.")
		return err
	} else {
		if err = ioutil.WriteFile(configPath, b, 0755); err != nil {
			log.Error(err)
			return err
		}
	}
	return nil
}
