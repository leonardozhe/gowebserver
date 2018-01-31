package models

import (
	// "strings"
	"bufio"
	log "github.com/cihub/seelog"
	//"github.com/fsnotify/fsnotify"
	"io"
	"os"
	"strings"
	"sync"
)

type SimpCamp struct {
	Uid string
	Url string
	Geo string
}

// XXX: TODO
// Shall create a common process for all of these models
//
var _simpcampConf = "./uid.conf"
var _simpcampMutex sync.RWMutex
var _simpcampCache = make(map[string]SimpCamp)

var _simpcampIPConf = "./blackip.conf"
var _simpcampIPMutex sync.RWMutex
var _simpcampIPCache = make(map[string]bool)
var _simpcampIPChan = make(chan string)

var _simpcampWhiteIPConf = "./whiteip.ake"
var _simpcampWhiteIPMutex sync.RWMutex
var _simpcampWhiteIPList = make(map[string]bool)

func _doSyncSimpCampCache(newCache map[string]SimpCamp) map[string]SimpCamp {
	_simpcampMutex.Lock()
	defer _simpcampMutex.Unlock()

	var tmp = _simpcampCache
	_simpcampCache = newCache
	return tmp
}

func syncSimpCamps() {
	fp, err := os.Open(_simpcampConf)
	if err != nil {
		log.Errorf("Failed to open %s error: %s", _simpcampConf, err)
		return
	}
	defer fp.Close()

	camps := make([]SimpCamp, 0, 100)
	br := bufio.NewReader(fp)
	for {
		line, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}

		tk := strings.Split(strings.TrimSpace(string(line)), "|")
		camps = append(camps, SimpCamp{Uid: tk[0], Url: tk[1], Geo: tk[2]})
	}

	var cache = make(map[string]SimpCamp)
	for _, v := range camps {
		cache[v.Uid] = v
	}

	var cachebak = _doSyncSimpCampCache(cache)
	for k, _ := range cachebak {
		delete(cachebak, k)
	}
	cachebak = nil
}

func GetSimCampConfs() []string {
	return []string{_simpcampConf, _simpcampWhiteIPConf}
}

func GetSimCamp(uid string) (SimpCamp, bool) {
	_simpcampMutex.Lock()
	defer _simpcampMutex.Unlock()

	c, ok := _simpcampCache[uid]
	return c, ok
}

func _doSyncSimpCampIPCache(newCache map[string]bool) map[string]bool {
	_simpcampIPMutex.Lock()
	defer _simpcampIPMutex.Unlock()

	var tmp = _simpcampIPCache
	_simpcampIPCache = newCache
	return tmp
}

func SyncSimpCamp(name string) {
	if name == _simpcampConf {
		syncSimpCamps()
	} else if name == _simpcampWhiteIPConf {
		initSimpCampWhiteIPs()
	}
}

func StartWatchSimpCampService() {
	syncSimpCamps()

	// watcher, err := fsnotify.NewWatcher()
	// if err != nil {
	// 	log.Critical(err)
	// }
	// defer watcher.Close()

	// go func() {
	// 	for {
	// 		select {
	// 		case event := <-watcher.Events:
	// 			log.Info("event:", event)
	// 			if event.Op&fsnotify.Write == fsnotify.Write {
	// 				syncSimpCamps()
	// 				log.Info("modified file:", event.Name)
	// 			}
	// 		case err := <-watcher.Errors:
	// 			log.Error("error:", err)
	// 		}
	// 	}
	// }()

	// // err = watcher.Add(_simpcampConf)
	// err = watcher.Add("/tmp/foo")
	// if err != nil {
	// 	log.Critical(err)
	// }
}

func _doSyncSimpCampWhiteIPs(newCache map[string]bool) map[string]bool {
	_simpcampWhiteIPMutex.Lock()
	defer _simpcampWhiteIPMutex.Unlock()

	var tmp = _simpcampWhiteIPList
	_simpcampWhiteIPList = newCache
	return tmp
}

func initSimpCampWhiteIPs() {
	fp, err := os.Open(_simpcampWhiteIPConf)
	if err != nil {
		return
	}
	defer fp.Close()

	ips := make([]string, 0, 1024*8)
	br := bufio.NewReader(fp)
	for {
		line, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		ips = append(ips, strings.TrimSpace(string(line)))
	}

	cache := make(map[string]bool)
	for _, ip := range ips {
		cache[ip] = true
	}

	cachebak := _doSyncSimpCampWhiteIPs(cache)
	for k, _ := range cachebak {
		delete(cachebak, k)
	}
	cachebak = nil
}

func IsWhiteIP(ip string) bool {
	_simpcampWhiteIPMutex.Lock()
	defer _simpcampWhiteIPMutex.Unlock()
	_, ok := _simpcampWhiteIPList[ip]
	return ok
}

func initSimpCampIPs() {
	fp, err := os.Open(_simpcampIPConf)
	if err != nil {
		log.Errorf("Failed to open %s error: %s", _simpcampIPConf, err)
		return
	}
	defer fp.Close()

	ips := make([]string, 0, 1024*8)
	br := bufio.NewReader(fp)
	for {
		line, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		ips = append(ips, strings.TrimSpace(string(line)))
	}

	cache := make(map[string]bool)
	for _, ip := range ips {
		cache[ip] = true
	}

	cachebak := _doSyncSimpCampIPCache(cache)
	for k, _ := range cachebak {
		delete(cachebak, k)
	}
	cachebak = nil
}

func SaveSimpCampIP(ip string) {
	_simpcampIPMutex.Lock()
	defer _simpcampIPMutex.Unlock()

	if _, ok := _simpcampIPCache[ip]; ok {
		return
	}

	_simpcampIPCache[ip] = true
	_simpcampIPChan <- ip

}

func InSimpCampIPBlackList(ip string) bool {
	_simpcampIPMutex.Lock()
	defer _simpcampIPMutex.Unlock()
	_, ok := _simpcampIPCache[ip]
	return ok
}

func apendSimpCampIPFile(ip string) {
	fp, err := os.OpenFile(_simpcampIPConf, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Error("Failed to open blackiplist.txt error: ", err)
		return
	}
	defer fp.Close()

	_, err = fp.WriteString(ip + "\n")
	if err != nil {
		log.Error("failed writing to blackiplist.txt: %s", err)
		return
	}
}

func StartSimpCampIPFileService() {
	initSimpCampIPs()
	initSimpCampWhiteIPs()

	go func() {
		for {
			select {
			case ip := <-_simpcampIPChan:
				// log.Info("add black ip ", ip)
				apendSimpCampIPFile(ip)
			}
		}
	}()
}
