package main

import (
	"crypto"
	"encoding/binary"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mkideal/log"
	"github.com/weikaishio/distributed_lib/buffer"

	"github.com/weikaishio/sim/codec"
	"github.com/weikaishio/sim/common/netutil"
)

const (
	MachineStatus_Connected = 0
	MachineStatus_HandShake = 1
	MachineStatus_Logined   = 2
	MachineStatus_Started   = 4
)

type Client struct {
	Id            uint32
	Key           string
	session       *netutil.RWSession
	bufPool       *buffer.Pool
	seq           uint32
	msgId         uint32
	running       int32 //处理杀进程信号所用
	runningUpload chan struct{}
	runningHeart  chan struct{}
	currentStatus int //当前机器状态
	lock          sync.Mutex
	handledSeq    string
	version       byte
}

func (m *Client) makeNewSeq() uint16 {
	val := atomic.AddUint32(&m.seq, 1)
	return uint16(val)
}
func (m *Client) makeNewMsgId() uint32 {
	val := atomic.AddUint32(&m.msgId, 1)
	return val
}
func newClient(id uint32, key string) *Client {
	return &Client{
		Id:            id,
		Key:           key,
		currentStatus: MachineStatus_Connected,
		bufPool:       buffer.NewPool(bufPoolSize, bufMaxPoolSize),
		runningUpload: make(chan struct{}),
		runningHeart:  make(chan struct{}),
		running:       1,
		lock:          sync.Mutex{},
		handledSeq:    "",
		version:       1,
	}
}
func (m *Client) onRecvPacket(packet []byte) {
	msg, err := codec.Decode(packet)
	if err != nil {
		//m.session.Quit(false)
		log.Error("onRecvPacket Decode err:%s", err)
		return
	}
	m.dispatchProto(msg)
}
func (m *Client) dispatchProto(msg interface{}) {
	switch ptc := msg.(type) {
	case *codec.HandShakeRes:
		m.onHandShakeRes(ptc)
	case *codec.AuthRes:
		m.onAuthRes(ptc)
	case *codec.HeartBeatRes:
		log.Info("HeartBeatRes ptc:%v", ptc)
	case *codec.SendMsgRes:
		m.onSendMsgRes(ptc)
	case *codec.RevMsg:
		m.onRevMsg(ptc)
	case *codec.ReadMsgRes:
		m.onReadMsgRes(ptc)
	default:
		log.Warn("unknow ptc:%v", ptc)
	}
}
func (m *Client) onRevMsg(ptc *codec.RevMsg) {
	log.Info("onRevMsg ptc:%v", ptc)

	proto := &codec.RevMsgRes{
		Cmd:    codec.CMD_RevMsgRes,
		Seq:    ptc.Seq,
		Status: codec.RESP_Status_Success,
	}
	m.sendProto(proto.Cmd, proto)

	m.readMsg(ptc.MsgId)
}
func (m *Client) readMsg(msgId uint32) {
	log.Info("readMsg msgId:%v", msgId)

	proto := &codec.RevMsgRes{
		Cmd:    codec.CMD_ReadMsg,
		Seq:    m.makeNewSeq(),
		Status: codec.RESP_Status_Success,
	}
	m.sendProto(proto.Cmd, proto)
}
func (m *Client) onReadMsgRes(ptc *codec.ReadMsgRes) {
	log.Info("onReadMsgRes ptc:%v", ptc)
}
func (m *Client) onSendMsgRes(ptc *codec.SendMsgRes) {
	log.Info("onSendMsgRes ptc:%v", ptc)
}
func ComputeHash(version byte, machineId, timestamp uint32, key string) (token []byte, err error) {
	cryptoSHA1 := crypto.SHA1
	hashSHA1 := cryptoSHA1.New()

	if _, err := hashSHA1.Write([]byte{version}); err != nil {
		return nil, err
	}
	buf4 := make([]byte, 4)
	binary.BigEndian.PutUint32(buf4, machineId)
	if _, err := hashSHA1.Write(buf4); err != nil {
		return nil, err
	}

	binary.BigEndian.PutUint32(buf4, timestamp)
	if _, err := hashSHA1.Write(buf4); err != nil {
		return nil, err
	}
	if _, err := hashSHA1.Write([]byte(key)); err != nil {
		return nil, err
	}
	hashed := hashSHA1.Sum(nil)
	return hashed, nil
}
func (m *Client) onHandShakeRes(ptc *codec.HandShakeRes) {
	log.Info("onHandShakeRes ptc:%v", ptc)

	timeStamp := time.Now().Unix()
	proto := &codec.Auth{
		Cmd: codec.CMD_Auth, Seq: uint16(m.makeNewSeq()),
		MachineId: 1,
		Timestamp: uint32(timeStamp),
	}
	token, err := ComputeHash(m.version, proto.MachineId, proto.Timestamp, m.Key)
	if err != nil {
		log.Error("ConputeHash proto:%v, err:%v", proto, err)
		return
	}
	proto.Token = token
	m.sendProto(proto.Cmd, proto)
	log.Info("sendAuth:%v", proto)
	m.currentStatus |= MachineStatus_HandShake
}
func (m *Client) onBeginHandle() {
	log.Info("onBeginHandle")
	proto := &codec.HandShake{
		Cmd:           codec.CMD_HandShake,
		Seq:           uint16(m.makeNewSeq()),
		ClientVersion: uint16(1),
		AppId:         m.Id,
	}
	m.sendProto(proto.Cmd, proto)
	//todo: exchange key by tsl
	log.Info("sendHandShake:%v", proto)
}
func (m *Client) onAuthRes(ptc *codec.AuthRes) {
	log.Info("onAuthRes ptc:%v", ptc)
	if ptc.Status == codec.RESP_Status_Fail {
		m.session.Quit(false)
		log.Warn("auth fail")
	} else {
		if (m.currentStatus & MachineStatus_Logined) == 0 {
			log.Info("(m.currentStatus & MachineStatus_Logined) == 0")
			go m.sendHeartbeat()
			go m.sendMsg()
		}

		m.handledSeq = ""
		m.currentStatus = m.currentStatus | MachineStatus_Logined
		log.Info("auth success")
	}
}
func (m *Client) sendProto(cmd uint16, proto interface{}) {
	log.Info("sendProto:%d,proto:%v begin", cmd, proto)
	packet, err := codec.Encode(cmd, proto)
	if err != nil {
		log.Warn("sendProto Encode:%v, err:%v", proto, err)
		return
	}
	buf, isFromPool := m.bufPool.Get()
	defer m.bufPool.Put(buf, isFromPool)
	_, err = buf.Write(packet)
	if err != nil {
		log.Warn("sendProto:%v,buf.Write err:%v", proto, err)
		return
	}
	m.session.Send(buf)
	log.Info("sendProto:%d,proto:%v end", cmd, proto)
}
func (m *Client) sendBeginHandle() {
	for m.isRunning() {
		select {
		case <-time.After(3600 * time.Second):
			m.onBeginHandle()
		}
	}
}
func (m *Client) sendHeartbeat() {
	if !m.isRunning(){
		log.Error("sendHeartbeat !m.isRunning()")
	}
	tick:=time.Tick(30*time.Second)
	for m.isRunning() {
		select {
		case <-tick:
			proto := &codec.HeartBeat{Cmd: codec.CMD_HeartBeat, Seq: m.makeNewSeq()}
			m.sendProto(proto.Cmd, proto)
		case <-m.runningHeart:
			log.Warn("stop heartbeat")
			break
		}
	}
}
func (m *Client) sendMsg() {
	for m.isRunning() {
		tick:=time.Tick(30*time.Second)
		select {
		case <-tick:

			proto := &codec.SendMsg{
				Cmd:     codec.CMD_SendMsg,
				Seq:     m.makeNewSeq(),
				MsgId:   m.makeNewMsgId(),
				Content: []byte("from client"),
			}
			m.sendProto(proto.Cmd, proto)
			log.Info("SendMsg:%v", proto)

		case <-m.runningUpload:
			log.Warn("stop dataupload")
			break
		}
	}
}
func (m *Client) stop() {
	m.runningUpload <- struct{}{}
	m.runningHeart <- struct{}{}
	atomic.CompareAndSwapInt32(&m.running, 1, 0)
}
func (m *Client) start() {
	if (m.currentStatus & MachineStatus_Started) > 0 {
		m.currentStatus = m.currentStatus - MachineStatus_Started
	}
	if (m.currentStatus & MachineStatus_Logined) > 0 {
		m.currentStatus = m.currentStatus - MachineStatus_Logined
	}
	atomic.StoreInt32(&m.running, 1)
}
func (m *Client) isRunning() bool {
	return atomic.LoadInt32(&m.running) != 0
}
