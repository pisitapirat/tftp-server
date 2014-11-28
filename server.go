package main

import (
	"net"
	"bytes"
	"fmt"
	"time"
)

type Server struct {
    Filesystem map[string][]byte
}

func udpListener() (*net.UDPConn, error) {
    addr, e := net.ResolveUDPAddr("udp", ":0")
    if e != nil {
        return nil, e
    }
    return net.ListenUDP("udp", addr)
}

func readPacket(clientConn *net.UDPConn, buffer []byte) (Packet, *net.UDPAddr, error) {
	e := clientConn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	if e != nil {
		return nil, nil, fmt.Errorf("Couldn't set timeout: %v", e)
	}
    n, remoteAddr, e := clientConn.ReadFromUDP(buffer)
    if e != nil {
        return nil, remoteAddr, e
    }
    p, e := ParsePacket(buffer[:n])
    if e != nil {
        return nil, remoteAddr, e
    }
    return p, remoteAddr, nil
}

func (s *Server) handleWrqPacket(p *WrqPacket, remoteAddr *net.UDPAddr) {
    clientConn, _ := udpListener()
    defer clientConn.Close()
    buffer := make([]byte, MAX_DATAGRAM_SIZE)

   	data := &bytes.Buffer{}
   	fmt.Printf("Saving file %v, mode %v\n", p.Filename, p.Filemode)
   	i := uint16(0)
   	reading := true

   	var fatalError error
   	retries := 0

	loop:
    for {
    	ack := &AckPacket{i}
	    clientConn.WriteToUDP(ack.Pack(), remoteAddr)

	    if !reading {
	    	break
	    }

        p, newAddr, e := readPacket(clientConn, buffer)
        if newAddr != nil {
        	remoteAddr = newAddr
        }
        if e != nil {
        	netError := e.(net.Error)
        	if netError.Timeout() {
        		// Try this block again
        		if retries > 4 {
        			fatalError = e
        			break
        		}
        		fmt.Printf("Timeout occurred, retrying\n")
        		retries += 1
        		continue
        	}
            fatalError = e
            break
        }

        switch p := p.(type) {
        case *DataPacket:
            if p.BlockNum != i + 1 {
                // Just ignore
                continue              
            }

            data.Write(p.Data)
            i += 1
            retries = 0

	        if len(p.Data) < MAX_BLOCK_SIZE {
	            reading = false
	        }

        default:
            fatalError = fmt.Errorf("Invalid packet received")
            break loop
        }
	}

	if fatalError != nil {
		err := &ErrorPacket{0, "An error occurred"}
        clientConn.WriteToUDP(err.Pack(), remoteAddr)
	} else {
		s.Filesystem[p.Filename] = data.Bytes()	
	}
    return
}

func (s *Server) handleRrqPacket(p *RrqPacket, remoteAddr *net.UDPAddr) {
    clientConn, e := udpListener()
    defer clientConn.Close()
    if e != nil {
        return
    }
    buffer := make([]byte, MAX_DATAGRAM_SIZE)

    fmt.Printf("Reading file %v, mode %v\n", p.Filename, p.Filemode)
    var fatalError error

    val, ok := s.Filesystem[p.Filename]
    if !ok {
	    fmt.Printf("'%v' wasn't found in filesystem\n", p.Filename)
	    notFoundErr := &ErrorPacket{1, "File not found"}
	    clientConn.WriteToUDP(notFoundErr.Pack(), remoteAddr)
	    return
    }
    fmt.Printf("%v is in filesystem, serving\n", p.Filename)

    i := 1
    retries := 0
    loop:
    for {
    	data := getBlock(val, i)
        dataPacket := &DataPacket{uint16(i), data}
        clientConn.WriteToUDP(dataPacket.Pack(), remoteAddr)

        p, newAddr, e := readPacket(clientConn, buffer)
        if newAddr != nil {
        	remoteAddr = newAddr
        }
        if e != nil {
        	netError := e.(net.Error)
        	if netError.Timeout() {
        		// Try this block again
        		if retries > 4 {
        			fatalError = e
        			break
        		}
        		fmt.Printf("Timeout occurred, retrying\n")
        		retries += 1
        		continue
        	}
        	fatalError = e
        	break
        }
        switch p := p.(type) {
        case *AckPacket:
        	// Ignore packets of other block numbers
        	if p.BlockNum == uint16(i) {
            	i += 1
            	retries = 0
            }
        case *ErrorPacket:
            fatalError = fmt.Errorf("An error occurred in the client: %d %v", p.ErrorCode, p.ErrMsg)
            break loop
        default:
            fatalError = fmt.Errorf("Invalid packet returned")
            break loop
        }
    }

    if fatalError != nil {
		err := &ErrorPacket{0, "An error occurred"}
        clientConn.WriteToUDP(err.Pack(), remoteAddr)
    }

    return
}

func (s *Server) ProcessRequest(conn *net.UDPConn) error {
	var buffer []byte
	buffer = make([]byte, MAX_DATAGRAM_SIZE)
	n, remoteAddr, e := conn.ReadFromUDP(buffer)

	packet, e := ParsePacket(buffer[:n])
    if e != nil {
        return e
    }

    switch p := packet.(type) {
    case *RrqPacket:
        go s.handleRrqPacket(p, remoteAddr)
    case *WrqPacket:
        go s.handleWrqPacket(p, remoteAddr)
    default:
        fmt.Printf("Invalid packet detected")
    }

	return nil
}

// Retrieve the part of the file for the given block number
func getBlock(contents []byte, i int) []byte {
    sliceStart := MAX_BLOCK_SIZE * (i-1)
    sliceEnd := sliceStart + MAX_BLOCK_SIZE
    if (sliceStart > len(contents)) {
        // Termination
        return nil
    } else if sliceEnd > len(contents) {
        sliceEnd = len(contents)
    }

    return contents[sliceStart:sliceEnd]	
}