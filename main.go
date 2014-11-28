package main

import (
	"fmt"
	"net"
)

func main() {
	addr, e := net.ResolveUDPAddr("udp", ":9229")
    fmt.Printf("Listening on :9229\n")
	if e != nil {
		fmt.Printf("Error resolving address: %v\n", e)
	}
	for {
		conn, e := net.ListenUDP("udp", addr)
		if e != nil {
			fmt.Printf("Error binding to udp address: %v\n", e)
		}
        fs := make(map[string][]byte, 0)
        server := &Server{fs}
		for {
			e = server.ProcessRequest(conn)
			if e != nil {
				fmt.Printf("Error processing request: %v\n", e)
			}
		}
	}
}
