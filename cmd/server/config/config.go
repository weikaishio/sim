package config

import (
	"github.com/mkideal/log"
	robfigconf "github.com/robfig/config"

	"github.com/weikaishio/sim/common"
)

type Config struct {
	Env      string
	Location string
	Dir      string

	LogDir   string
	LogLevel string

	Host          string
	Port          int
	KeepaliveTime int

	HttpPort     int //健康监控所用
	IsProduction bool
	EncryptKey   string
}

func NewConfig(dir, env, location string) (*Config, error) {
	cfg := new(Config)

	cfg.Env = env
	cfg.Dir = dir
	cfg.Location = location

	return cfg, nil
}

func (cfg *Config) Clone() (*Config, error) {
	cfgClone := &Config{
		Env:      cfg.Env,
		Location: cfg.Location,
		Dir:      cfg.Dir,

		LogDir:   cfg.LogDir,
		LogLevel: cfg.LogLevel,

		Host:          cfg.Host,
		Port:          cfg.Port,
		KeepaliveTime: cfg.KeepaliveTime,

		HttpPort:     cfg.HttpPort,
		IsProduction: cfg.IsProduction,
		EncryptKey:   cfg.EncryptKey,
	}
	return cfgClone, nil
}

// config reload需要把之前的配置克隆，如果新配置加载失败继续使用之前的配置
func (cfg *Config) Reload() (*Config, error) {
	cfgClone, err := cfg.Clone()
	if err != nil {
		log.Error("cfg.reload clone err:%v", err)
		return cfg, err
	}

	basic := cfg.Dir + "basic.conf"
	err = cfg.loadBasicConfig(basic)
	if err != nil {
		return cfgClone, err
	}

	advanced := cfg.Dir + cfg.Location + cfg.Env + ".conf"
	err = cfg.loadAdvancedConfig(advanced)
	if err != nil {
		return cfg, err
	} else {
		return cfg, nil
	}
}
func (cfg *Config) loadBasicConfig(conf string) error {
	c, err := robfigconf.ReadDefault(conf)
	if err != nil {
		return err
	}

	// 日志配置
	cfg.LogDir, cfg.LogLevel = common.ReadLogConfig(c, false)

	cfg.Host, cfg.Port, cfg.KeepaliveTime = common.ReadNetConfig("net", c, false)

	// web服务配置
	cfg.HttpPort, cfg.EncryptKey, cfg.IsProduction = common.ReadWebConfig(c, false)

	return nil
}
func (cfg *Config) loadAdvancedConfig(conf string) error {
	c, err := robfigconf.ReadDefault(conf)
	if err != nil {
		return err
	}

	// 重写web服务配置
	httpPort, encryptKey, isProduction := common.ReadWebConfig(c, true)
	if httpPort > 0 && encryptKey != "" {
		cfg.HttpPort = httpPort
		cfg.EncryptKey = encryptKey
		cfg.IsProduction = isProduction
	}
	// 重写日志配置
	logDir, logLevel := common.ReadLogConfig(c, true)
	if logDir != "" && logLevel != "" {
		cfg.LogDir = logDir
		cfg.LogLevel = logLevel
	}
	host, port, keepaliveTime := common.ReadNetConfig("net", c, true)
	if host != "" && port > 0 && keepaliveTime > 0 {
		cfg.Host = host
		cfg.Port = port
		cfg.KeepaliveTime = keepaliveTime
	}
	return nil
}
