package codec

//todo:首包识别，magic
//todo:协议升级，新增字段？header+pb?

//握手协议
//cmd=0x0
type HandShake struct {
	Cmd           uint16
	Seq           uint16
	ClientVersion uint16
	AppId         uint32
}

type HandShakeRes struct {
	Cmd         uint16
	Seq         uint16
	Status      byte
	EncryptType uint8
}

//认证协议
//cmd=0x1
type Auth struct {
	Cmd       uint16
	Seq       uint16
	MachineId uint32
	Timestamp uint32
	Token     []byte //count=20
}

type AuthRes struct {
	Cmd    uint16
	Seq    uint16
	Status byte
}

//心跳协议
//cmd=0x2
type HeartBeat struct {
	Cmd uint16
	Seq uint16
}

type HeartBeatRes struct {
	Cmd    uint16
	Seq    uint16
	Status byte
}

//cmd=0x3
type SendMsg struct {
	Cmd     uint16
	Seq     uint16
	MsgId   uint32
	Content []byte
}

type SendMsgRes struct {
	Cmd    uint16
	Seq    uint16
	Status byte
}

//cmd=0x4
type RevMsg struct {
	Cmd   uint16
	Seq   uint16
	MsgId uint32
}

type RevMsgRes struct {
	Cmd    uint16
	Seq    uint16
	Status byte
}

//cmd=0x5
type ReadMsg struct {
	Cmd   uint16
	Seq   uint16
	MsgId uint32
}

type ReadMsgRes struct {
	Cmd    uint16
	Seq    uint16
	Status byte
}