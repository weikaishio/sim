package config

import (
	"errors"
	"flag"
	"path"

	"github.com/weikaishio/sim/common/confutil"

	"github.com/BurntSushi/toml"
)

var (
	// Conf global config variable
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
	flag.StringVar(&confPath, "configDir", "./config", "config path")
	flag.StringVar(&PidFile, "pid", "sim_gate.pid", "pid filepath")
}

//Init int config
func Init() error {
	if confPath != "" {
		path := path.Join(confPath, "/conf.toml")
		return local(path)
	}
	return errors.New("confPath is nil")
}

func local(path string) (err error) {
	_, err = toml.DecodeFile(path, &Conf)
	return
}
