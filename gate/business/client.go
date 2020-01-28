package business

import "github.com/weikaishio/sim/common/netutil"

type Client interface {
	GetId() uint32
	GetUniqueId() uint32
	GetSession()  *netutil.RWSession
}
