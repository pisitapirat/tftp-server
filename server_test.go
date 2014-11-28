package main

import (
	"fmt"
	"net"
	"testing"
)

func TestServer(t *testing.T) {

	go func() {
        fs := make(map[string][]byte, 0)
        server := &Server{fs}
        s.handleRrqPacket()

		addr, e := net.ResolveUDPAddr("udp", ":9229")	
		conn, e := net.ListenUDP("udp", addr)
	}
	clientConn := udpListener()
	clientConn.WriteToUDP(notFoundErr.Pack(), remoteAddr)
}