package config

import (
	"errors"
	"flag"

	"github.com/weikaishio/sim/common/confutil"

	"github.com/BurntSushi/toml"
)

var (
	Conf     = &Config{}
	confPath string
	PidFile  string
)

type Config struct {
	Net       *confutil.TcpConf
	Redis     *confutil.Options
	Log       *confutil.LogConfig
	GrpcImCli *confutil.GRPCCliConf
	Etcd      *confutil.Etcd
}

func init() {
	flag.StringVar(&confPath, "configDir", "./config/conf.toml", "config path")
	flag.StringVar(&PidFile, "pid", "sim_gate.pid", "pid filepath")
}

func Init() error {
	if confPath != "" {
		_, err := toml.DecodeFile(confPath, &Conf)
		return err
	}
	return errors.New("confPath is nil")
}
