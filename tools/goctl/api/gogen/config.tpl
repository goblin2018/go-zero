package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	DB struct {
		DataSource string
	}
	Mongo struct {
		Url string
	}
	Redis struct {
		Host string
		Type string
	}
	Auth  AuthConf
	Cache cache.CacheConf
}

type AuthConf struct {
	Key    string
	Expire int64
}
