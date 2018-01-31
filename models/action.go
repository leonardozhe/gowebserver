package models

import (
	"os"
	"path"
	"sync"

	// Driver for postgres
	"github.com/leonardozhe/gowebserver/utils"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/lib/pq"
)

var _ip2locMutex sync.RWMutex

// InitMapDB to initialize the db of ip2location
func InitLocDB() {

	var filepath string

	currentPath, _ := os.Getwd()
	filepath = path.Join(path.Dir(currentPath),
		"gowebserver/models/mapdb/ip.bin")

	utils.Ip2LocInit(filepath)
}

func GetIp2LocInfos(ip string) utils.IP2Locationrecord {
	_simpcampWhiteIPMutex.Lock()
	defer _simpcampWhiteIPMutex.Unlock()
	return utils.Get_CountryShort_UsageType(ip)
}
