package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/leonardozhe/gowebserver/configfile"
	"github.com/leonardozhe/gowebserver/handlers"
	"github.com/leonardozhe/gowebserver/models"
	log "github.com/cihub/seelog"
	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/olebedev/config"
)

var PORT string = ""

func initLog() {
	logger, err := log.LoggerFromConfigAsString(configfile.LogConfig)
	if err != nil {
		log.Error("log fail to start. logging is impossible.", err)
	}
	log.ReplaceLogger(logger)
}

func startServer() {
	fmt.Println("server starting")
	conf, err := config.ParseYaml(configfile.ConfString)
	if err != nil {
		log.Critical("config fail to load, no way to start the server")
	}
	PORT, err = conf.String("port")
	if err != nil {
		PORT = "3000"
	}

	r := gin.Default()

	r.GET("/robot.txt", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "User-agent: *\nDisallow: /")
		return
	})

	r.GET("stat/view", handlers.ServeURLHandler)

	fmt.Println("server started")
	r.Run(":" + PORT)
}

func main() {
	cpus := runtime.NumCPU()
	runtime.GOMAXPROCS(cpus)

	models.InitLocDB()

	err := models.InitRedisPool()
	if err != nil {
		log.Error(err)
	}

	initLog()

	// ips := []string{"27.102.70.170", "218.88.126.133",
	// 	"47.52.145.205"}
	// log.Infof("========Start-1========\n")
	// idx := 0
	// for i := 0; i < 10000; i++ {
	// 	for _, ip := range ips {
	// 		idx += 1
	// 		go func(ip string) {
	// 			ipInfos := models.GetIp2LocInfos(ip)
	// 			log.Infof("%v, %s, country %s, usageType %s", idx, ip,
	// 				ipInfos.Country_short, ipInfos.Usagetype)
	// 		}(ip)
	// 	}
	// }
	// log.Infof("========End========\n")
	// done := make(chan bool)
	// <-done

	// defer func() {
	// 	if r := recover(); r != nil {
	// 		log.Error("Recovered in f", r)
	// 	}
	// }()
	// defer log.Flush()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-c
		// log.Debug("system end ", utils.SystemEnd)
		log.Info("Stopping. signal: ", s)
		// Do something.
		log.Info("Stopped")
		os.Exit(0)
	}()

	models.StartWatchSimpCampService()
	models.StartSimpCampIPFileService()

	fmt.Println("preparation finished")
	//models.DB.LogMode(false)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Critical(err)
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				// log.Info("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					models.SyncSimpCamp(event.Name)
					log.Info("modified file:", event.Name)
				}
			case err := <-watcher.Errors:
				log.Error("error:", err)
			}
		}
	}()

	for _, v := range models.GetSimCampConfs() {
		if err := watcher.Add(v); err != nil {
			log.Critical(err)
		}
	}

	startServer()

}
