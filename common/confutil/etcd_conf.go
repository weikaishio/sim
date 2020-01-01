package confutil

import "strings"

type Etcd struct {
	Endpoints string
	Username  string
	Password  string
}

func (e *Etcd) EndpointAry() []string {
	return strings.Split(e.Endpoints, ",")
}
