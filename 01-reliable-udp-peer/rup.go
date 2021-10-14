package rup

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type ReliableUdpPeer struct {
	logger         *log.Logger
	name           string
	ReceivedBuffer []string
	inFlightLimit  chan interface{}
	wg             sync.WaitGroup
}

func NewReliableUdpPeer(logger *log.Logger, peerName string) *ReliableUdpPeer {
	p := &ReliableUdpPeer{logger, peerName, []string{}, make(chan interface{}, 3), sync.WaitGroup{}}
	return p

}

func (p *ReliableUdpPeer) Listen(address string) {
	data := make([]byte, 2048)

	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		p.Log("%v", err)
		return
	}

	ln, err := net.ListenUDP("udp", addr)
	if err != nil {
		p.Log("%v", err)
		return
	}

	go func() {
		defer ln.Close()

		for {
			length, remoteAddr, err := ln.ReadFromUDP(data)
			if err != nil {
				p.Log("%v", err)
				return
			}

			p.ReceivedBuffer = append(p.ReceivedBuffer, string(data[:length]))
			p.Log("From %v: %s", remoteAddr, data[:length])

			_, err = ln.WriteToUDP([]byte("ack"), remoteAddr)
			if err != nil {
				p.Log("%v", err)
			}
		}
	}()
}

func (p *ReliableUdpPeer) Send(conn *net.UDPConn, destAddr *net.UDPAddr, msg string) (err error) {
	_, err = conn.WriteToUDP([]byte(msg), destAddr)
	if err != nil {
		return
	}

	data := make([]byte, 2048)
	deadline := time.Now().Add(time.Second)
	err = conn.SetReadDeadline(deadline)
	if err != nil {
		return
	}

	_, remoteAddr, err := conn.ReadFromUDP(data)
	if err != nil {
		return
	}

	p.Log("%s acked", remoteAddr)
	return
}

func (p *ReliableUdpPeer) ReliableSend(msg string, address string) {
	destAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		p.Log("%v", err)
		return
	}

	conn, err := net.ListenUDP("udp", nil)
	if err != nil {
		p.Log("%v", err)
		return
	}

	go func() {
		defer conn.Close()
		p.wg.Add(1)

		p.inFlightLimit <- nil

		for {
			err = p.Send(conn, destAddr, msg)

			if err != nil {
				p.Log("%v", err)
			} else {
				<-p.inFlightLimit
				p.wg.Done()
				return
			}
		}
	}()
}

func (p *ReliableUdpPeer) Wait() {
	p.wg.Wait()
}

func (p *ReliableUdpPeer) Log(msg string, a ...interface{}) {
	p.logger.Printf("Peer %v: %v \n", p.name, fmt.Sprintf(msg, a...))
}
