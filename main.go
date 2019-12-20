package main

import (
	"flag"
	"log"
	"net"
)

func main() {
	localAddr := flag.String("local", ":1700", "local address to tunnel from")
	remoteAddr := flag.String("remote", "127.0.0.1:1701", "remote address to tunnel to")
	flag.Parse()

	la, err := net.ResolveUDPAddr("udp", *localAddr)
	if err != nil {
		log.Fatalf("Failed to resolve local address: %s", err)
	}
	l, err := net.ListenUDP("udp", la)
	if err != nil {
		log.Fatalf("Failed to listen on local address: %s", err)
	}

	conns := make(map[string]net.Conn)
	b := make([]byte, 65507)
	for {
		n, addr, err := l.ReadFromUDP(b)
		if err != nil {
			log.Printf("Failed to read UDP packet on local address: %s", err)
			continue
		}
		conn, ok := conns[addr.String()]
		if !ok {
			conn, err = net.Dial("udp", *remoteAddr)
			if err != nil {
				log.Printf("Failed to dial %s: %s", *remoteAddr, err)
				continue
			}
			conns[addr.String()] = conn
			go func() {
				b := make([]byte, 65507)
				for {
					n, err := conn.Read(b)
					if err != nil {
						log.Printf("Failed to read packet from %s: %s", conn.RemoteAddr(), err)
						continue
					}
					_, err = l.WriteToUDP(b[:n], addr)
					if err != nil {
						log.Printf("Failed to write packet to %s: %s", addr, err)
					}
				}
			}()
		}
		_, err = conn.Write(b[:n])
		if err != nil {
			log.Printf("Failed to write packet to %s: %s", conn.RemoteAddr(), err)
		}
	}
}
