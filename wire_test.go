package main

import (
	"bytes"
	"testing"
)

func TestErrorPacket(t *testing.T) {
	errPacket := &ErrorPacket{1, "Test error"}
	data := errPacket.Pack()
	p, _ := ParsePacket(data)
	switch p := p.(type) {
	case *ErrorPacket:
		if p.ErrorCode != 1 {
			t.Errorf("Invalid error code, should be 1")
		}

		if p.ErrMsg != "Test error" {
			t.Errorf("Invalid error message: %v\n", p.ErrMsg)
		}
	default:
		t.Errorf("Packet type isnt correct\n")
	}
}

func TestAckPacket(t *testing.T) {
	ackPacket := &AckPacket{123}
	data := ackPacket.Pack()
	p, _ := ParsePacket(data)
	switch p := p.(type) {
	case *AckPacket:
		if p.BlockNum != 123 {
			t.Errorf("Invalid block number")
		}
	default:
		t.Errorf("Packet type isnt correct\n")
	}
}

func TestDataPacket(t *testing.T) {
	blockData := []byte{0x1, 0x2, 0x3}

	dataPacket := &DataPacket{123, blockData}
	data := dataPacket.Pack()
	p, _ := ParsePacket(data)
	switch p := p.(type) {
	case *DataPacket:
		if p.BlockNum != 123 {
			t.Errorf("Invalid block number")
		}

		if bytes.Compare(p.Data, blockData) != 0 {
			t.Errorf("Not correct data")
		}
	default:
		t.Errorf("Packet type isnt correct\n")
	}
}

func TestRqPacket(t *testing.T) {
	wrqPacket := &WrqPacket{"test_file.txt", "octet"}
	data := wrqPacket.Pack()
	p, _ := ParsePacket(data)
	switch p := p.(type) {
	case *WrqPacket:
		if p.Filename != "test_file.txt" {
			t.Errorf("Invalid filename: %v\n", p.Filename)
		}

		if p.Filemode != "octet" {
			t.Errorf("Invalid filemode: %v\n", p.Filemode)
		}
	default:
		t.Errorf("Packet type isnt correct\n")
	}
}
