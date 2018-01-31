package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	log "github.com/cihub/seelog"
	"github.com/garyburd/redigo/redis"
	"github.com/olebedev/config"

	"github.com/leonardozhe/gowebserver/configfile"
)

var (
	redisPool                redis.Pool
	hasRedis                 bool
	ErrorMissingRedisAddress = errors.New("redis: server address not found")
	ErrorRedisNotAvailable   = errors.New("redis: server not available")
)

func GetRedisConn() redis.Conn {
	if hasRedis {
		return redisPool.Get()
	}

	return nil
}

func RedisGetStruct(key string, out interface{}) error {
	var err error

	value := reflect.ValueOf(out)
	if value.Kind() != reflect.Ptr || value.IsNil() {
		return errors.New(fmt.Sprint(value, " is not a valid pointer "))
	}
	if value.Elem().Kind() != reflect.Struct {
		return errors.New(fmt.Sprint(value, " is not a struct pointer"))
	}

	conn := GetRedisConn()
	if conn == nil {
		return ErrorRedisNotAvailable
	}
	defer conn.Close()
	byts, err := redis.Bytes(conn.Do("GET", key))
	if err != nil || len(byts) == 0 {
		return ErrorRedisNotAvailable
	}

	if err = json.Unmarshal(byts, out); err != nil {

		return err
	}

	return nil
}

func RedisSetStruct(key string, out interface{}) error {
	return RedisSetStructEx(key, out, -1)
}
func RedisSetStructEx(key string, out interface{}, timeout int) error {
	var err error

	value := reflect.ValueOf(out)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return errors.New(fmt.Sprint(reflect.TypeOf(out),
			" is not a struct pointer"))
	}

	conn := GetRedisConn()
	if conn == nil {
		return ErrorRedisNotAvailable
	}
	defer conn.Close()

	str, err := json.Marshal(out)
	if err != nil {
		return err
	}

	//XXX: Do we need to move the hard code expiration to config file?
	if timeout < 0 {
		_, err = conn.Do("SET", key, str)
	} else {
		_, err = conn.Do("SETEX", key, timeout, str)
	}

	if err != nil {

		return err
	}

	return nil
}

func RedisDelStruct(key string) error {
	var err error

	conn := GetRedisConn()
	if conn == nil {
		return ErrorRedisNotAvailable
	}
	defer conn.Close()

	if _, err = conn.Do("DEL", key); err != nil {
		return err
	}

	return nil
}

func RedisDelKeys(keyPattern string) error {
	conn := GetRedisConn()
	defer conn.Close()

	keys, err := redis.Strings(conn.Do("KEYS", keyPattern))
	if err != nil {
		return err
	}

	for _, v := range keys {
		conn.Do("DEL", v)
	}

	return nil
}

func InitRedisPool() error {
	conf, err := config.ParseYaml(configfile.ConfString)
	if err != nil {
		log.Critical(err)

		return err
	}

	url, err := conf.String("redis.url")
	if err != nil {
		log.Critical(err)

		return err
	}

	index, err := conf.Int("redis.index")
	if err != nil {
		log.Critical(err)

		return err
	}

	poolMaxIdle, _ := conf.Int("redis.pool_max_idle")
	if poolMaxIdle > 0 {
		redisPool.MaxIdle = poolMaxIdle
	}

	poolMaxActive, _ := conf.Int("redis.pool_max_active")
	if poolMaxActive > 0 {
		redisPool.MaxActive = poolMaxActive
	}

	poolIdleTimeOut, _ := conf.Int("redis.pool_idle_time_out")
	if poolIdleTimeOut > 0 {
		redisPool.IdleTimeout = time.Duration(poolIdleTimeOut) * time.Second
	}

	redisPool.Dial = func() (redis.Conn, error) {
		// c, err := redis.DialURL(url, redis.DialDatabase(index), redis.DialPassword("AAbb1122"))
		c, err := redis.DialURL(url, redis.DialDatabase(index))
		if err != nil {
			log.Warn("Failed to connect to redis server. Error:",
				err)

			return nil, err
		}

		return c, nil
	}

	hasRedis = true

	return nil
}

func HasRedis() bool {
	return hasRedis
}
