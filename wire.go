package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
)

const (
	OPCODE_RRQ = uint16(1)
	OPCODE_WRQ = uint16(2)
	OPCODE_DAT = uint16(3)
	OPCODE_ACK = uint16(4)
	OPCODE_ERR = uint16(5)
)

const (
	MAX_BLOCK_SIZE        = 512
	MAX_DATAGRAM_SIZE = 516 // 512 bytes data + opcode + block#
)

type Packet interface {
	Unpack(data []byte) error
	Pack() []byte
}

/*
	WRQ/RRQ:

	opcode (2 bytes)
	filename (string)
	0 (1 byte)
	mode (string) -- only support "octet"
	0 (1 byte)
*/

type RrqPacket struct {
	Filename string
	Filemode string // Should always be octet
}

func (p *RrqPacket) Unpack(data []byte) (e error) {
	p.Filename, p.Filemode, e = unpackRqPacket(data)
	return
}

func (p *RrqPacket) Pack() []byte {
	return packRqPacket(p.Filename, p.Filemode, OPCODE_WRQ)
}

type WrqPacket struct {
	Filename string
	Filemode string // Should always be octet
}

func (p *WrqPacket) Unpack(data []byte) (e error) {
	p.Filename, p.Filemode, e = unpackRqPacket(data)
	return
}

func (p *WrqPacket) Pack() []byte {
	return packRqPacket(p.Filename, p.Filemode, OPCODE_WRQ)
}

func unpackRqPacket(data []byte) (filename string, filemode string, e error) {
	buffer := bytes.NewBuffer(data[2:])
	filename, e = buffer.ReadString(0x0)
	if e != nil {
		return
	}
	filename = strings.TrimRight(filename, "\x00")
	filemode, e = buffer.ReadString(0x0)
	filemode = strings.TrimRight(filemode, "\x00")
	return
}

func packRqPacket(filename string, filemode string, opcode uint16) []byte {
	buffer := &bytes.Buffer{}
	binary.Write(buffer, binary.BigEndian, opcode)
	buffer.WriteString(filename)
	buffer.WriteByte(0x0)
	buffer.WriteString(filemode)
	buffer.WriteByte(0x0)
	return buffer.Bytes()
}

/*
	DataPacket (DATA):

	Opcode (2 bytes)
	block # (2 bytes)
	data (n bytes)
*/

type DataPacket struct {
	BlockNum uint16
	Data     []byte
}

func (p *DataPacket) Unpack(data []byte) error {
	p.BlockNum = binary.BigEndian.Uint16(data[2:])
	p.Data = data[4:]
	return nil
}

func (p *DataPacket) Pack() []byte {
	buffer := &bytes.Buffer{}
	binary.Write(buffer, binary.BigEndian, OPCODE_DAT)
	binary.Write(buffer, binary.BigEndian, p.BlockNum)
	buffer.Write(p.Data)
	return buffer.Bytes()
}

/*
	AckPacket (ACK):

	Opcode (2 bytes)
	Block # (2 bytes)
*/

type AckPacket struct {
	BlockNum uint16
}

func (p *AckPacket) Unpack(data []byte) error {
	p.BlockNum = binary.BigEndian.Uint16(data[2:])
	return nil
}

func (p *AckPacket) Pack() []byte {
	buffer := &bytes.Buffer{}
	binary.Write(buffer, binary.BigEndian, OPCODE_ACK)
	binary.Write(buffer, binary.BigEndian, p.BlockNum)
	return buffer.Bytes()
}

/*
	ErrorPacket (ERROR)

	Opcode (2 bytes)
	ErrorCode (2 bytes)
	ErrMsg (string)
	0 (1 byte)

    0         Not defined, see error message (if any).
    1         File not found.
    2         Access violation.
    3         Disk full or allocation exceeded.
    4         Illegal TFTP operation.
    5         Unknown transfer ID.
    6         File already exists.
    7         No such user.
*/

type ErrorPacket struct {
	ErrorCode uint16
	ErrMsg    string
}

func (p *ErrorPacket) Unpack(data []byte) error {
	p.ErrorCode = binary.BigEndian.Uint16(data[2:])
	buffer := bytes.NewBuffer(data[4:])
	errorMsg, e := buffer.ReadString(0x0)
	if e != nil {
		return e
	}
	p.ErrMsg = strings.TrimRight(errorMsg, "\x00")
	return nil
}

func (p *ErrorPacket) Pack() []byte {
	buffer := &bytes.Buffer{}
	binary.Write(buffer, binary.BigEndian, OPCODE_ERR)

	// tftp on OSX is actually wanting this to be LittleEndian,
	// I dont know why
	binary.Write(buffer, binary.LittleEndian, p.ErrorCode)
	buffer.WriteString(p.ErrMsg)
	buffer.WriteByte(0x0)
	return buffer.Bytes()
}

func ParsePacket(bytes []byte) (Packet, error) {
	var p Packet
	opcode := binary.BigEndian.Uint16(bytes)
	switch opcode {
	case OPCODE_RRQ:
		p = &RrqPacket{}
	case OPCODE_WRQ:
		p = &WrqPacket{}
	case OPCODE_DAT:
		p = &DataPacket{}
	case OPCODE_ACK:
		p = &AckPacket{}
	case OPCODE_ERR:
		p = &ErrorPacket{}
	default:
		return nil, fmt.Errorf("Invalid packet opcode: %d", opcode)
	}

	e := p.Unpack(bytes)
	if e != nil {
		return nil, e
	}
	return p, nil
}
