package worker

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Worker struct {
	IP        string
	Port      int
	Timer     *time.Timer
	Free      float64
	Total     float64
	Usage     float64 // percent of used
	InfoValid bool
}

func (w *Worker) GetCapInfo() {
	w.InfoValid = false
	if rsp, err := http.Get("http://" + w.IP + ":" + strconv.Itoa(w.Port) + "/system/info"); err != nil {
		log.Println(err)
	} else {
		if (rsp.StatusCode/100 == 4) || (rsp.StatusCode/100 == 5) {
			log.Printf("server return %d.", rsp.StatusCode)
		} else {
			if buf, err := ioutil.ReadAll(rsp.Body); err != nil {
				log.Println(err)
			} else {
				msg := make(map[string]interface{})
				if err := json.Unmarshal(buf, &msg); err != nil {
					log.Println(err, string(buf))
				} else {
					if "running" == msg["status"].(string) {
						w.Free = msg["free"].(float64)
						w.Total = msg["total"].(float64)
						w.Usage = (w.Total - w.Free) / w.Total
						w.InfoValid = true
					} else {
						log.Printf("failed to get cap info, reason:%s\n", msg["detail"])
					}
				}
			}
		}
	}
}

type WorkerList struct {
	List []*Worker
}

func (wl WorkerList) String() string {
	var tmp string
	for i, w := range wl.List {
		tmp += strconv.Itoa(i) + " " + w.IP + ":" + strconv.Itoa(w.Port) + " " + strconv.FormatFloat(w.Usage, 'f', -1, 64) + "\n"
	}
	return tmp
}

func (wl WorkerList) Len() int {
	return len(wl.List)
}

func (wl WorkerList) Swap(a, b int) {
	wl.List[a], wl.List[b] = wl.List[b], wl.List[a]
}

func (wl WorkerList) Less(a, b int) bool {
	return wl.List[a].Usage < wl.List[b].Usage
}
