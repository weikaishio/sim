#!/bin/sh

#//cmd=0xC0  cmd命令ID
#type SendMsg struct {
#	Cmd     uint16
#	Seq     uint16
#	MsgId   uint32
#   From    string   //len=30
#   From    []byte   //count=30
#	Content []byte   //不定长度的必须放在最后一个，否则需要指定 count或者len
#}
#
#type SendMsgRes struct {
#	Cmd    uint16
#	Seq    uint16
#	Status byte
#}
#

./codecgen

go fmt ./codec.go