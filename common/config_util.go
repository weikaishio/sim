package common

import (
	"errors"
	"fmt"
	"github.com/mkideal/log"
	robfigconf "github.com/robfig/config"
	"runtime"
	"strings"
)

func ReadLogConfig(c *robfigconf.Config, rewrite bool) (path, level string) {
	section := "log"
	if !rewrite || c.HasOption(section, "path") {
		path, _ = c.String(section, "path")
	}
	if !rewrite || c.HasOption(section, "level") {
		level, _ = c.String(section, "level")
	}
	return
}
func ReadWebConfig(c *robfigconf.Config, rewrite bool) (httpPort int, encrypKey string, isPruduction bool) {
	section := "web"
	if !rewrite || c.HasOption(section, "port") {
		httpPort, _ = c.Int(section, "port")
	}
	if !rewrite || c.HasOption(section, "encryptKey") {
		encrypKey, _ = c.String(section, "encryptKey")
	}
	if !rewrite || c.HasOption(section, "isProduction") {
		isPruduction, _ = c.Bool(section, "isProduction")
	}
	return
}
func ReadNetConfig(section string, c *robfigconf.Config, rewrite bool) (host string, port, keepAliveTime int) {
	if !rewrite || c.HasOption(section, "host") {
		host, _ = c.String(section, "host")
	}
	if !rewrite || c.HasOption(section, "port") {
		port, _ = c.Int(section, "port")
	}
	if !rewrite || c.HasOption(section, "keepalive-time") {
		keepAliveTime, _ = c.Int(section, "keepalive-time")
	}
	return
}

func ReadEtcdConfig(c *robfigconf.Config) (endpoints []string, username, password string) {
	section := "etcd"
	endpointStr, _ := c.String(section, "endpoints")
	endpoints = strings.Split(endpointStr, ",")
	username, _ = c.String(section, "username")
	password, _ = c.String(section, "password")
	return
}
func ReadGRPCConfig(c *robfigconf.Config) (key string, port int, host string, isProd bool, queueType int, certFile, keyFile string, isShowRedisOrmLog, isShowMysqlOrmLog bool) {
	section := "grpc"
	port, _ = c.Int(section, "port")
	host, _ = c.String(section, "host")
	key, _ = c.String(section, "key")
	isProd, _ = c.Bool(section, "is_prod")
	queueType, _ = c.Int(section, "queueType")

	isShowRedisOrmLog, _ = c.Bool(section, "isShowRedisOrmLog")
	isShowMysqlOrmLog, _ = c.Bool(section, "isShowMysqlOrmLog")

	certFile, _ = c.String(section, "cert_file")
	keyFile, _ = c.String(section, "key_file")
	return
}

func TryExec(fn func()) (err error) {
	defer func() {
		if e := recover(); e != nil {
			buf := make([]byte, 1<<16)
			buf = buf[:runtime.Stack(buf, true)]
			switch typ := e.(type) {
			case error:
				err = typ
			case string:
				err = errors.New(typ)
			default:
				err = fmt.Errorf("%v", typ)
			}
			log.Error("==== STACK TRACE BEGIN ====\npanic: %v\n%s\n===== STACK TRACE END =====", err, string(buf))
		}
	}()
	fn()
	return
}
