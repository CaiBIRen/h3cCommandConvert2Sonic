package redisclient

import (
	"context"
	"fmt"

	"github.com/coreos/pkg/capnslog"
	"github.com/go-redis/redis/v8"
)

var REDISCLIENT *redis.Client
var glog = capnslog.NewPackageLogger("sonic-unis-framework", "REDISCLIENT")

func NewClient() {
	REDISCLIENT = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis服务器地址
		Password: "",               // 默认情况下不需要密码
		DB:       15,               // 使用db15
	})
	_, err := REDISCLIENT.Ping(context.TODO()).Result()
	if err != nil {
		panic("redis connecting error")
	}
}

func IndexSet(key string, value int) {
	_, err := REDISCLIENT.Set(context.TODO(), key, value, 0).Result()
	if err != nil {
		fmt.Println("set redis key err", err)
	}
	glog.Infof("key %s,value %d set success", key, value)
}

func IndexGet(key string) (string, error) {
	value, err := REDISCLIENT.Get(context.TODO(), key).Result()
	if err != nil {
		glog.Errorf("redis get key %s err %v", key, err)
		return "", err
	}
	glog.Infof("key %s  get success", key)
	return value, nil
}
