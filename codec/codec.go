package codec

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"runtime"

	"github.com/mkideal/log"
)

const (
	LEN_MSGLEN = 2
	LEN_MSGCMD = 2
	LEN_MSGSEQ = 2
)
const (
	RESP_Status_Success = 0x0
	RESP_Status_Fail    = 0x1
)
const (
	CMD_HandShake    = 0x0
	CMD_HandShakeRes = 0x80
	CMD_Auth         = 0x1
	CMD_AuthRes      = 0x81
	CMD_HeartBeat    = 0x2
	CMD_HeartBeatRes = 0x82
	CMD_SendMsg      = 0xC0
	CMD_SendMsgRes   = 0x40
	CMD_RevMsg       = 0x3
	CMD_RevMsgRes    = 0x83
)

var (
	MAGIC     = "cw"
	LEN_MAGIC = len(MAGIC)

	ERR_NOTIMPLEMETN = errors.New("not implement")
	ERR_LEN          = errors.New("packet length error")
	ERR_MAGIC        = errors.New("packet error")
)

func Decode(buf []byte) (obj interface{}, err error) {
	defer func() {
		if e := recover(); e != nil {
			var err error
			buf := make([]byte, 1<<16)
			buf = buf[:runtime.Stack(buf, true)]
			switch typ := e.(type) {
			case error:
				err = typ
			case string:
				err = errors.New(typ)
			default:
				err = fmt.Errorf("%v", typ)
			}
			log.Error("==== STACK TRACE BEGIN ====\npanic: %v\n%s\n===== STACK TRACE END =====", err, string(buf))
		}
	}()
	baseLen := LEN_MAGIC + LEN_MSGLEN + LEN_MSGCMD + LEN_MSGSEQ
	if len(buf) < baseLen {
		return nil, errors.New(fmt.Sprintf("packet is invalid:%v", buf))
	}
	magic, cmd, seq, bodyLen := DecodeHeader(buf)
	if string(magic) != MAGIC {
		return nil, ERR_MAGIC
	}

	switch cmd {

	case CMD_HandShake:
		obj, err := DecodeHandShake(buf[baseLen:], bodyLen)
		if err != nil {
			return nil, err
		}
		obj.Seq = seq
		obj.Cmd = cmd
		return obj, err

	case CMD_HandShakeRes:
		obj, err := DecodeHandShakeRes(buf[baseLen:], bodyLen)
		if err != nil {
			return nil, err
		}
		obj.Seq = seq
		obj.Cmd = cmd
		return obj, err

	case CMD_Auth:
		obj, err := DecodeAuth(buf[baseLen:], bodyLen)
		if err != nil {
			return nil, err
		}
		obj.Seq = seq
		obj.Cmd = cmd
		return obj, err

	case CMD_AuthRes:
		obj, err := DecodeAuthRes(buf[baseLen:], bodyLen)
		if err != nil {
			return nil, err
		}
		obj.Seq = seq
		obj.Cmd = cmd
		return obj, err

	case CMD_HeartBeat:
		obj, err := DecodeHeartBeat(buf[baseLen:], bodyLen)
		if err != nil {
			return nil, err
		}
		obj.Seq = seq
		obj.Cmd = cmd
		return obj, err

	case CMD_HeartBeatRes:
		obj, err := DecodeHeartBeatRes(buf[baseLen:], bodyLen)
		if err != nil {
			return nil, err
		}
		obj.Seq = seq
		obj.Cmd = cmd
		return obj, err

	case CMD_SendMsg:
		obj, err := DecodeSendMsg(buf[baseLen:], bodyLen)
		if err != nil {
			return nil, err
		}
		obj.Seq = seq
		obj.Cmd = cmd
		return obj, err

	case CMD_SendMsgRes:
		obj, err := DecodeSendMsgRes(buf[baseLen:], bodyLen)
		if err != nil {
			return nil, err
		}
		obj.Seq = seq
		obj.Cmd = cmd
		return obj, err

	case CMD_RevMsg:
		obj, err := DecodeRevMsg(buf[baseLen:], bodyLen)
		if err != nil {
			return nil, err
		}
		obj.Seq = seq
		obj.Cmd = cmd
		return obj, err

	case CMD_RevMsgRes:
		obj, err := DecodeRevMsgRes(buf[baseLen:], bodyLen)
		if err != nil {
			return nil, err
		}
		obj.Seq = seq
		obj.Cmd = cmd
		return obj, err

	}

	return nil, ERR_NOTIMPLEMETN
}

func Encode(cmd uint16, obj interface{}) ([]byte, error) {
	defer func() {
		if e := recover(); e != nil {
			var err error
			buf := make([]byte, 1<<16)
			buf = buf[:runtime.Stack(buf, true)]
			switch typ := e.(type) {
			case error:
				err = typ
			case string:
				err = errors.New(typ)
			default:
				err = fmt.Errorf("%v", typ)
			}
			log.Error("==== STACK TRACE BEGIN ====\npanic: %v\n%s\n===== STACK TRACE END =====", err, string(buf))
		}
	}()

	switch cmd {

	case CMD_HandShake:
		return EncodeHandShake(obj.(*HandShake))

	case CMD_HandShakeRes:
		return EncodeHandShakeRes(obj.(*HandShakeRes))

	case CMD_Auth:
		return EncodeAuth(obj.(*Auth))

	case CMD_AuthRes:
		return EncodeAuthRes(obj.(*AuthRes))

	case CMD_HeartBeat:
		return EncodeHeartBeat(obj.(*HeartBeat))

	case CMD_HeartBeatRes:
		return EncodeHeartBeatRes(obj.(*HeartBeatRes))

	case CMD_SendMsg:
		return EncodeSendMsg(obj.(*SendMsg))

	case CMD_SendMsgRes:
		return EncodeSendMsgRes(obj.(*SendMsgRes))

	case CMD_RevMsg:
		return EncodeRevMsg(obj.(*RevMsg))

	case CMD_RevMsgRes:
		return EncodeRevMsgRes(obj.(*RevMsgRes))

	}
	return nil, ERR_NOTIMPLEMETN
}

func DecodeHeader(buf []byte) (magic []byte, cmd, seq, bodyLen uint16) {
	offset := 0
	magic = buf[offset : offset+LEN_MAGIC]
	offset += LEN_MAGIC
	len := binary.BigEndian.Uint16(buf[offset:])
	bodyLen = len - uint16(LEN_MAGIC) - LEN_MSGLEN - LEN_MSGCMD - LEN_MSGSEQ
	offset += LEN_MSGLEN
	cmd = binary.BigEndian.Uint16(buf[offset:])
	offset += 2
	seq = binary.BigEndian.Uint16(buf[offset:])
	return
}
func EncodeHeader(cmd, bodyLen, seq uint16) ([]byte, error) {
	packetBuf := bytes.NewBufferString("")

	packetLen := uint16(LEN_MAGIC) + LEN_MSGLEN + LEN_MSGCMD + LEN_MSGSEQ + bodyLen

	if _, err := packetBuf.Write([]byte(MAGIC)); err != nil {
		return nil, err
	}
	buf2 := make([]byte, 2)
	binary.BigEndian.PutUint16(buf2, uint16(packetLen))
	if _, err := packetBuf.Write(buf2); err != nil {
		return nil, err
	}

	cmdBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(cmdBuf, cmd)
	if _, err := packetBuf.Write(cmdBuf); err != nil {
		return nil, err
	}

	seqBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(seqBuf, seq)
	if _, err := packetBuf.Write(seqBuf); err != nil {
		return nil, err
	}
	return packetBuf.Bytes(), nil
}

func EncodeHandShake(handShake *HandShake) ([]byte, error) {
	packetBuf := bytes.NewBufferString("")

	bodyLen := 10
	if packetHeader, err := EncodeHeader(handShake.Cmd, uint16(bodyLen), handShake.Seq); err != nil {
		return nil, err
	} else if _, err := packetBuf.Write(packetHeader); err != nil {
		return nil, err
	}

	bufCmd := make([]byte, 2)
	binary.BigEndian.PutUint16(bufCmd, handShake.Cmd)
	if _, err := packetBuf.Write(bufCmd); err != nil {
		return nil, err
	}
	bufSeq := make([]byte, 2)
	binary.BigEndian.PutUint16(bufSeq, handShake.Seq)
	if _, err := packetBuf.Write(bufSeq); err != nil {
		return nil, err
	}
	bufClientVersion := make([]byte, 2)
	binary.BigEndian.PutUint16(bufClientVersion, handShake.ClientVersion)
	if _, err := packetBuf.Write(bufClientVersion); err != nil {
		return nil, err
	}
	bufAppId := make([]byte, 4)
	binary.BigEndian.PutUint32(bufAppId, handShake.AppId)
	if _, err := packetBuf.Write(bufAppId); err != nil {
		return nil, err
	}

	return packetBuf.Bytes(), nil
}
func DecodeHandShake(buf []byte, bodyLen uint16) (handShake *HandShake, err error) {
	if uint16(len(buf)) != bodyLen || bodyLen < 10 {
		return nil, ERR_LEN
	}
	handShake = &HandShake{}
	offset := 0

	handShake.Cmd = binary.BigEndian.Uint16(buf[offset:])
	offset += 2
	handShake.Seq = binary.BigEndian.Uint16(buf[offset:])
	offset += 2
	handShake.ClientVersion = binary.BigEndian.Uint16(buf[offset:])
	offset += 2
	handShake.AppId = binary.BigEndian.Uint32(buf[offset:])
	offset += 4

	return
}

func EncodeHandShakeRes(handShakeRes *HandShakeRes) ([]byte, error) {
	packetBuf := bytes.NewBufferString("")

	bodyLen := 6
	if packetHeader, err := EncodeHeader(handShakeRes.Cmd, uint16(bodyLen), handShakeRes.Seq); err != nil {
		return nil, err
	} else if _, err := packetBuf.Write(packetHeader); err != nil {
		return nil, err
	}

	bufCmd := make([]byte, 2)
	binary.BigEndian.PutUint16(bufCmd, handShakeRes.Cmd)
	if _, err := packetBuf.Write(bufCmd); err != nil {
		return nil, err
	}
	bufSeq := make([]byte, 2)
	binary.BigEndian.PutUint16(bufSeq, handShakeRes.Seq)
	if _, err := packetBuf.Write(bufSeq); err != nil {
		return nil, err
	}
	if _, err := packetBuf.Write([]byte{handShakeRes.Status}); err != nil {
		return nil, err
	}
	if _, err := packetBuf.Write([]byte{handShakeRes.EncryptType}); err != nil {
		return nil, err
	}

	return packetBuf.Bytes(), nil
}
func DecodeHandShakeRes(buf []byte, bodyLen uint16) (handShakeRes *HandShakeRes, err error) {
	if uint16(len(buf)) != bodyLen || bodyLen < 6 {
		return nil, ERR_LEN
	}
	handShakeRes = &HandShakeRes{}
	offset := 0

	handShakeRes.Cmd = binary.BigEndian.Uint16(buf[offset:])
	offset += 2
	handShakeRes.Seq = binary.BigEndian.Uint16(buf[offset:])
	offset += 2
	handShakeRes.Status = buf[offset]
	offset += 1
	handShakeRes.EncryptType = buf[offset]
	offset += 1

	return
}

func EncodeAuth(auth *Auth) ([]byte, error) {
	packetBuf := bytes.NewBufferString("")

	bodyLen := 32
	if packetHeader, err := EncodeHeader(auth.Cmd, uint16(bodyLen), auth.Seq); err != nil {
		return nil, err
	} else if _, err := packetBuf.Write(packetHeader); err != nil {
		return nil, err
	}

	bufCmd := make([]byte, 2)
	binary.BigEndian.PutUint16(bufCmd, auth.Cmd)
	if _, err := packetBuf.Write(bufCmd); err != nil {
		return nil, err
	}
	bufSeq := make([]byte, 2)
	binary.BigEndian.PutUint16(bufSeq, auth.Seq)
	if _, err := packetBuf.Write(bufSeq); err != nil {
		return nil, err
	}
	bufMachineId := make([]byte, 4)
	binary.BigEndian.PutUint32(bufMachineId, auth.MachineId)
	if _, err := packetBuf.Write(bufMachineId); err != nil {
		return nil, err
	}
	bufTimestamp := make([]byte, 4)
	binary.BigEndian.PutUint32(bufTimestamp, auth.Timestamp)
	if _, err := packetBuf.Write(bufTimestamp); err != nil {
		return nil, err
	}

	if _, err := packetBuf.Write(auth.Token); err != nil {
		return nil, err
	}

	return packetBuf.Bytes(), nil
}
func DecodeAuth(buf []byte, bodyLen uint16) (auth *Auth, err error) {
	if uint16(len(buf)) != bodyLen || bodyLen < 32 {
		return nil, ERR_LEN
	}
	auth = &Auth{}
	offset := 0

	auth.Cmd = binary.BigEndian.Uint16(buf[offset:])
	offset += 2
	auth.Seq = binary.BigEndian.Uint16(buf[offset:])
	offset += 2
	auth.MachineId = binary.BigEndian.Uint32(buf[offset:])
	offset += 4
	auth.Timestamp = binary.BigEndian.Uint32(buf[offset:])
	offset += 4
	auth.Token = buf[offset : offset+20]
	offset += 20

	return
}

func EncodeAuthRes(authRes *AuthRes) ([]byte, error) {
	packetBuf := bytes.NewBufferString("")

	bodyLen := 5
	if packetHeader, err := EncodeHeader(authRes.Cmd, uint16(bodyLen), authRes.Seq); err != nil {
		return nil, err
	} else if _, err := packetBuf.Write(packetHeader); err != nil {
		return nil, err
	}

	bufCmd := make([]byte, 2)
	binary.BigEndian.PutUint16(bufCmd, authRes.Cmd)
	if _, err := packetBuf.Write(bufCmd); err != nil {
		return nil, err
	}
	bufSeq := make([]byte, 2)
	binary.BigEndian.PutUint16(bufSeq, authRes.Seq)
	if _, err := packetBuf.Write(bufSeq); err != nil {
		return nil, err
	}
	if _, err := packetBuf.Write([]byte{authRes.Status}); err != nil {
		return nil, err
	}

	return packetBuf.Bytes(), nil
}
func DecodeAuthRes(buf []byte, bodyLen uint16) (authRes *AuthRes, err error) {
	if uint16(len(buf)) != bodyLen || bodyLen < 5 {
		return nil, ERR_LEN
	}
	authRes = &AuthRes{}
	offset := 0

	authRes.Cmd = binary.BigEndian.Uint16(buf[offset:])
	offset += 2
	authRes.Seq = binary.BigEndian.Uint16(buf[offset:])
	offset += 2
	authRes.Status = buf[offset]
	offset += 1

	return
}

func EncodeHeartBeat(heartBeat *HeartBeat) ([]byte, error) {
	packetBuf := bytes.NewBufferString("")

	bodyLen := 4
	if packetHeader, err := EncodeHeader(heartBeat.Cmd, uint16(bodyLen), heartBeat.Seq); err != nil {
		return nil, err
	} else if _, err := packetBuf.Write(packetHeader); err != nil {
		return nil, err
	}

	bufCmd := make([]byte, 2)
	binary.BigEndian.PutUint16(bufCmd, heartBeat.Cmd)
	if _, err := packetBuf.Write(bufCmd); err != nil {
		return nil, err
	}
	bufSeq := make([]byte, 2)
	binary.BigEndian.PutUint16(bufSeq, heartBeat.Seq)
	if _, err := packetBuf.Write(bufSeq); err != nil {
		return nil, err
	}

	return packetBuf.Bytes(), nil
}
func DecodeHeartBeat(buf []byte, bodyLen uint16) (heartBeat *HeartBeat, err error) {
	if uint16(len(buf)) != bodyLen || bodyLen < 4 {
		return nil, ERR_LEN
	}
	heartBeat = &HeartBeat{}
	offset := 0

	heartBeat.Cmd = binary.BigEndian.Uint16(buf[offset:])
	offset += 2
	heartBeat.Seq = binary.BigEndian.Uint16(buf[offset:])
	offset += 2

	return
}

func EncodeHeartBeatRes(heartBeatRes *HeartBeatRes) ([]byte, error) {
	packetBuf := bytes.NewBufferString("")

	bodyLen := 5
	if packetHeader, err := EncodeHeader(heartBeatRes.Cmd, uint16(bodyLen), heartBeatRes.Seq); err != nil {
		return nil, err
	} else if _, err := packetBuf.Write(packetHeader); err != nil {
		return nil, err
	}

	bufCmd := make([]byte, 2)
	binary.BigEndian.PutUint16(bufCmd, heartBeatRes.Cmd)
	if _, err := packetBuf.Write(bufCmd); err != nil {
		return nil, err
	}
	bufSeq := make([]byte, 2)
	binary.BigEndian.PutUint16(bufSeq, heartBeatRes.Seq)
	if _, err := packetBuf.Write(bufSeq); err != nil {
		return nil, err
	}
	if _, err := packetBuf.Write([]byte{heartBeatRes.Status}); err != nil {
		return nil, err
	}

	return packetBuf.Bytes(), nil
}
func DecodeHeartBeatRes(buf []byte, bodyLen uint16) (heartBeatRes *HeartBeatRes, err error) {
	if uint16(len(buf)) != bodyLen || bodyLen < 5 {
		return nil, ERR_LEN
	}
	heartBeatRes = &HeartBeatRes{}
	offset := 0

	heartBeatRes.Cmd = binary.BigEndian.Uint16(buf[offset:])
	offset += 2
	heartBeatRes.Seq = binary.BigEndian.Uint16(buf[offset:])
	offset += 2
	heartBeatRes.Status = buf[offset]
	offset += 1

	return
}

func EncodeSendMsg(sendMsg *SendMsg) ([]byte, error) {
	packetBuf := bytes.NewBufferString("")

	bodyLen := 8
	if packetHeader, err := EncodeHeader(sendMsg.Cmd, uint16(bodyLen), sendMsg.Seq); err != nil {
		return nil, err
	} else if _, err := packetBuf.Write(packetHeader); err != nil {
		return nil, err
	}

	bufCmd := make([]byte, 2)
	binary.BigEndian.PutUint16(bufCmd, sendMsg.Cmd)
	if _, err := packetBuf.Write(bufCmd); err != nil {
		return nil, err
	}
	bufSeq := make([]byte, 2)
	binary.BigEndian.PutUint16(bufSeq, sendMsg.Seq)
	if _, err := packetBuf.Write(bufSeq); err != nil {
		return nil, err
	}
	bufMsgId := make([]byte, 4)
	binary.BigEndian.PutUint32(bufMsgId, sendMsg.MsgId)
	if _, err := packetBuf.Write(bufMsgId); err != nil {
		return nil, err
	}

	if _, err := packetBuf.Write(sendMsg.Content); err != nil {
		return nil, err
	}

	return packetBuf.Bytes(), nil
}
func DecodeSendMsg(buf []byte, bodyLen uint16) (sendMsg *SendMsg, err error) {
	if uint16(len(buf)) != bodyLen || bodyLen < 8 {
		return nil, ERR_LEN
	}
	sendMsg = &SendMsg{}
	offset := 0

	sendMsg.Cmd = binary.BigEndian.Uint16(buf[offset:])
	offset += 2
	sendMsg.Seq = binary.BigEndian.Uint16(buf[offset:])
	offset += 2
	sendMsg.MsgId = binary.BigEndian.Uint32(buf[offset:])
	offset += 4
	sendMsg.Content = buf[offset:]

	return
}

func EncodeSendMsgRes(sendMsgRes *SendMsgRes) ([]byte, error) {
	packetBuf := bytes.NewBufferString("")

	bodyLen := 5
	if packetHeader, err := EncodeHeader(sendMsgRes.Cmd, uint16(bodyLen), sendMsgRes.Seq); err != nil {
		return nil, err
	} else if _, err := packetBuf.Write(packetHeader); err != nil {
		return nil, err
	}

	bufCmd := make([]byte, 2)
	binary.BigEndian.PutUint16(bufCmd, sendMsgRes.Cmd)
	if _, err := packetBuf.Write(bufCmd); err != nil {
		return nil, err
	}
	bufSeq := make([]byte, 2)
	binary.BigEndian.PutUint16(bufSeq, sendMsgRes.Seq)
	if _, err := packetBuf.Write(bufSeq); err != nil {
		return nil, err
	}
	if _, err := packetBuf.Write([]byte{sendMsgRes.Status}); err != nil {
		return nil, err
	}

	return packetBuf.Bytes(), nil
}
func DecodeSendMsgRes(buf []byte, bodyLen uint16) (sendMsgRes *SendMsgRes, err error) {
	if uint16(len(buf)) != bodyLen || bodyLen < 5 {
		return nil, ERR_LEN
	}
	sendMsgRes = &SendMsgRes{}
	offset := 0

	sendMsgRes.Cmd = binary.BigEndian.Uint16(buf[offset:])
	offset += 2
	sendMsgRes.Seq = binary.BigEndian.Uint16(buf[offset:])
	offset += 2
	sendMsgRes.Status = buf[offset]
	offset += 1

	return
}

func EncodeRevMsg(revMsg *RevMsg) ([]byte, error) {
	packetBuf := bytes.NewBufferString("")

	bodyLen := 8
	if packetHeader, err := EncodeHeader(revMsg.Cmd, uint16(bodyLen), revMsg.Seq); err != nil {
		return nil, err
	} else if _, err := packetBuf.Write(packetHeader); err != nil {
		return nil, err
	}

	bufCmd := make([]byte, 2)
	binary.BigEndian.PutUint16(bufCmd, revMsg.Cmd)
	if _, err := packetBuf.Write(bufCmd); err != nil {
		return nil, err
	}
	bufSeq := make([]byte, 2)
	binary.BigEndian.PutUint16(bufSeq, revMsg.Seq)
	if _, err := packetBuf.Write(bufSeq); err != nil {
		return nil, err
	}
	bufMsgId := make([]byte, 4)
	binary.BigEndian.PutUint32(bufMsgId, revMsg.MsgId)
	if _, err := packetBuf.Write(bufMsgId); err != nil {
		return nil, err
	}

	return packetBuf.Bytes(), nil
}
func DecodeRevMsg(buf []byte, bodyLen uint16) (revMsg *RevMsg, err error) {
	if uint16(len(buf)) != bodyLen || bodyLen < 8 {
		return nil, ERR_LEN
	}
	revMsg = &RevMsg{}
	offset := 0

	revMsg.Cmd = binary.BigEndian.Uint16(buf[offset:])
	offset += 2
	revMsg.Seq = binary.BigEndian.Uint16(buf[offset:])
	offset += 2
	revMsg.MsgId = binary.BigEndian.Uint32(buf[offset:])
	offset += 4

	return
}

func EncodeRevMsgRes(revMsgRes *RevMsgRes) ([]byte, error) {
	packetBuf := bytes.NewBufferString("")

	bodyLen := 5
	if packetHeader, err := EncodeHeader(revMsgRes.Cmd, uint16(bodyLen), revMsgRes.Seq); err != nil {
		return nil, err
	} else if _, err := packetBuf.Write(packetHeader); err != nil {
		return nil, err
	}

	bufCmd := make([]byte, 2)
	binary.BigEndian.PutUint16(bufCmd, revMsgRes.Cmd)
	if _, err := packetBuf.Write(bufCmd); err != nil {
		return nil, err
	}
	bufSeq := make([]byte, 2)
	binary.BigEndian.PutUint16(bufSeq, revMsgRes.Seq)
	if _, err := packetBuf.Write(bufSeq); err != nil {
		return nil, err
	}
	if _, err := packetBuf.Write([]byte{revMsgRes.Status}); err != nil {
		return nil, err
	}

	return packetBuf.Bytes(), nil
}
func DecodeRevMsgRes(buf []byte, bodyLen uint16) (revMsgRes *RevMsgRes, err error) {
	if uint16(len(buf)) != bodyLen || bodyLen < 5 {
		return nil, ERR_LEN
	}
	revMsgRes = &RevMsgRes{}
	offset := 0

	revMsgRes.Cmd = binary.BigEndian.Uint16(buf[offset:])
	offset += 2
	revMsgRes.Seq = binary.BigEndian.Uint16(buf[offset:])
	offset += 2
	revMsgRes.Status = buf[offset]
	offset += 1

	return
}
