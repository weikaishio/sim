package confutil

import "strings"

type GRPCSvrConf struct {
	Target   string
	Host     string
	Port     int
	CertFile string
	KeyFile  string
}

type GRPCCliConf struct {
	Target         string
	CertServerName string
	CertFile       string
	AuthName       string
	AuthPwd        string
}

type GRPCAuths struct {
	Auths string
}

func (g *GRPCAuths) AuthMap() map[string]string {
	authAry := strings.Split(g.Auths, "&")
	accounts := make(map[string]string)
	for _, item := range authAry {
		ap := strings.Split(item, "=")
		if len(ap) == 2 {
			accounts[ap[0]] = ap[1]
		}
	}
	return accounts
}
