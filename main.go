package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type IpList struct {
	sync.Mutex
	Ips []string
}

var (
	inport   = flag.Int("inport", 2626, "the port for incoming registration packets")
	bindaddr = flag.String("bindaddr", "0.0.0.0", "the address to listen on for incoming udp packets")
	outport  = flag.Int("outport", 2626, "the port on which to forward packets for all the hosts")
	dnsname  = flag.String("dnsname", "example.com", "the dns to resolve at each packet")
	ttl      = flag.Int("ttl", 60, "time in seconds to wait between each dns resolution")
	ips      = IpList{}
)

func gResolver(dns string, ttl int) {
	log.Printf("starting resolution goroutine for %s with ttl %d", dns, ttl)
	c := time.Tick(time.Duration(ttl) * time.Second)
	for _ = range c {
		ip, err := net.LookupHost(dns)
		if err != nil {
			log.Printf("Error resolving: %s: %s", dns, err)
		} else {
			log.Printf("Resolved %s as %s", dns, ip)
			ips.Lock()
			ips.Ips = ip
			ips.Unlock()
		}
	}
}

func gSendToAddress(addr string, data []byte) {
	udpaddr, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		log.Printf("cannot resolve %s: %s", addr, err)
		return
	}
	conn, err := net.DialUDP("udp4", nil, udpaddr)
	if err != nil {
		log.Printf("cannot DialUDP to %s: %s", addr, err)
		return
	}
	defer conn.Close()
	_, err = conn.Write(data)
	if err != nil {
		log.Printf("cannot Write to %s: %s", addr, err)
		return
	}
}

func gHandleUdpPacket(data []byte) {
	ips.Lock()
	for _, ip := range ips.Ips {
		go gSendToAddress(fmt.Sprintf("%s:%d", ip, *outport), data)
	}
	ips.Unlock()
}

func main() {
	flag.Parse()

	go gResolver(*dnsname, *ttl)

	laddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", *bindaddr, *inport))
	if err != nil {
		log.Printf("Cannot resolve %s:%d: %s", *bindaddr, *inport, err)
		return
	}
	conn, err := net.ListenUDP("udp4", laddr)
	if err != nil {
		log.Printf("Cannot listen on %s: %s", laddr, err)
		return
	}
	for {
		b := make([]byte, 1500)
		_, _, _, raddr, err := conn.ReadMsgUDP(b, nil)
		if err != nil {
			log.Printf("error reading udp message: %s", err)
			return
		}
		log.Printf("got packet from: %s", raddr)
		go gHandleUdpPacket(b)
	}
}
