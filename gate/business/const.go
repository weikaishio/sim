package business

const (
	bufPoolSize     = 100
	bufMaxPoolSize  = 1000
	maxConWriteSize = 1000

	maxMsgChanSize = 10000
)

const (
	WaitRespMaxTimeout = 15
	CommandTimeout     = 60
)

type ClientStatus int

const (
	ClientStatus_Connected  = 0
	ClientStatus_HandShaked = 1
	ClientStatus_Logined    = 2
)
