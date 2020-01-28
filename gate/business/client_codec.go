package business

import (
	"sync/atomic"
	"time"

	"github.com/weikaishio/sim/codec"
	"github.com/weikaishio/sim/common/netutil"
	"github.com/weikaishio/sim/gate/config"

	"github.com/mkideal/log"
)

type ClientCodec struct {
	uniqueId       uint32
	id             uint32
	svr            *Server
	lastActiveTime time.Time
	currentStatus  ClientStatus
	session        *netutil.RWSession
	seq            uint32
	version        byte
	remoteAddr     string
}

func (c *ClientCodec) makeNewSeq() uint16 {
	val := atomic.AddUint32(&c.seq, 1)
	return uint16(val)
}
func newClientCodec(s *Server, seq uint32, remoteAddr string) *ClientCodec {
	return &ClientCodec{
		uniqueId:       seq,
		svr:            s,
		remoteAddr:     remoteAddr,
		lastActiveTime: time.Now(),
		currentStatus:  ClientStatus_Connected,
	}
}
func (c *ClientCodec) GetId() uint32{
	return c.id
}
func (c *ClientCodec) GetUniqueId() uint32{
	return c.uniqueId
}
func (c *ClientCodec) GetSession()       *netutil.RWSession{
	return c.session
}
func (c *ClientCodec) onRecvPacket(packet []byte) {
	if !c.svr.isRunning() {
		log.Warn("client onRecvPacket:%d, svr is not running", c.id)
		return
	}
	now := time.Now()
	if int(now.Sub(c.lastActiveTime).Seconds()) > config.Conf.Net.KeepaliveTime {
		msg, err := codec.Decode(packet)
		log.Warn("Client onRecvPacket timeout:%d > %d,msg:%v,decodeErr:%v", int(now.Sub(c.lastActiveTime).Seconds()), config.Conf.Net.KeepaliveTime, msg, err)
		c.session.Quit(false)
		return
	}
	c.lastActiveTime = time.Now()
	//todo:magic recognition
	msg, err := codec.Decode(packet)
	if err != nil {
		log.Error("onRecvPacket Decode:%v err:%s", packet, err)
		return
	} else {
		log.Info("onRecvPacket mid:%d, msg:%v", c.id, msg)
	}
	c.dispatchProto(msg)
}
func (c *ClientCodec) dispatchProto(msg interface{}) {
	switch ptc := msg.(type) {
	case *codec.HandShake:
		c.onHandShake(ptc)
	case *codec.Auth:
		log.Info("ptc:%v", ptc)
		c.onAuth(ptc)
	case *codec.HeartBeat:
		c.onHeartbeat(ptc)
	case *codec.SendMsg:
		c.onSendMsg(ptc)
	case *codec.ReadMsg:
		c.onReadMsg(ptc)
	default:
		log.Warn("dispatchProto not support cmd:%v", ptc)
	}
}
func (c *ClientCodec) onSendMsg(ptc *codec.SendMsg) {
	log.Info("onSendMsg ptc:%v", ptc)
	res := &codec.SendMsgRes{
		Cmd:    codec.CMD_SendMsgRes,
		Seq:    ptc.Seq,
		Status: codec.RESP_Status_Success,
	}
	c.sendProto(res.Cmd, res)
}
func (c *ClientCodec) onReadMsg(ptc *codec.ReadMsg) {
	log.Info("onReadMsg ptc:%v", ptc)
	res := &codec.ReadMsgRes{
		Cmd:    codec.CMD_ReadMsgRes,
		Seq:    ptc.Seq,
		Status: codec.RESP_Status_Success,
	}
	c.sendProto(res.Cmd, res)
}
func (c *ClientCodec) onAuth(ptc *codec.Auth) {
	log.Info("onAuth:%d,:%v", c.id, ptc)
	authRes := &codec.AuthRes{
		Cmd:    codec.CMD_AuthRes,
		Seq:    ptc.Seq,
		Status: codec.RESP_Status_Success,
	}
	c.id = ptc.MachineId
	c.sendProto(authRes.Cmd, authRes)
	c.currentStatus |= ClientStatus_Logined

	go func() {
		for {
			if c.currentStatus&ClientStatus_Logined == 0 {
				break
			}
			time.Sleep(3 * time.Second)
		}
	}()
}
func (c *ClientCodec) onHandShake(ptc *codec.HandShake) {
	log.Info("onHandShake:%d,:%v", c.id, ptc)
	res := &codec.HandShakeRes{
		Cmd:    codec.CMD_HandShakeRes,
		Seq:    ptc.Seq,
		Status: codec.RESP_Status_Success,
	}
	c.sendProto(res.Cmd, res)
	c.currentStatus |= ClientStatus_HandShaked
}

func (c *ClientCodec) onHeartbeat(ptc *codec.HeartBeat) {
	log.Info("onHeartbeat:%d,:%v", c.id, ptc)
	if c.currentStatus&ClientStatus_Logined == 0 {
		log.Warn("machine:%d, not logined onHeartbeat:%v", c.id, ptc)
		proto := &codec.HeartBeatRes{Cmd: codec.CMD_HeartBeatRes, Seq: ptc.Seq, Status: codec.RESP_Status_Fail}
		c.sendProto(proto.Cmd, proto)
		return
	}
	proto := &codec.HeartBeatRes{Cmd: codec.CMD_HeartBeatRes, Seq: ptc.Seq, Status: codec.RESP_Status_Success}
	c.sendProto(proto.Cmd, proto)
}
func (c *ClientCodec) sendProto(cmd uint16, proto interface{}) {
	packet, err := codec.Encode(cmd, proto)
	if err != nil {
		log.Warn("sendProto Encode:%v, err:%v", proto, err)
		return
	}
	buf, isFromPool := c.svr.bufPool.Get()
	defer c.svr.bufPool.Put(buf, isFromPool)
	_, err = buf.Write(packet)
	if err != nil {
		log.Warn("sendProto:%v,buf.Write err:%v", proto, err)
		return
	}
	c.session.Send(buf)
}
func (c *ClientCodec) onNewSession() {
	log.Warn("onNewSession:%v", c)
}

func (c *ClientCodec) onQuitSession() {
	c.svr = nil
	c.session = nil
	c.currentStatus = ClientStatus_Connected
	log.Info("onQuitSession m:%v", c)
}
