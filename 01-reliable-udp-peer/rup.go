package rup

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

// ReliableUdpPeer represent a peer that
// can recieve and send messages reliable
// with ack message via udp protocol.
// it guarantee receive package to destination
// and limit number of inflight messages.
type ReliableUdpPeer struct {
	logger         *log.Logger      // for print result of each section in other goroutine
	name           string           // name of peer for better debuging
	ReceivedBuffer []string         // all recieived message are store in ReceivedBuffer
	inFlightLimit  chan interface{} // limit number of message without ack
	wg             sync.WaitGroup   // insure that all messages are received to the destination
}

// create ReliableUdpPeer and return it's pointer.
// logger handle the logging functionality.
// peerName is name of peer for debug info.
// listenAddr is address that peer listen for incoming messages.
// sendLimit number of messages that peer can send without recieved ack.
func NewReliableUdpPeer(logger *log.Logger, peerName string, listenAddr string, sendLimit int) (*ReliableUdpPeer, error) {
	// create ReliableUdpPeer
	p := &ReliableUdpPeer{logger, peerName, []string{}, make(chan interface{}, sendLimit), sync.WaitGroup{}}
	// lisnten on address
	err := p.listen(listenAddr)
	if err != nil {
		return nil, err
	}
	return p, nil

}

// create a socket for peer and wait for packet in
// a separate goroutine (light weight thread) a sotore
// received message in ReceivedBuffer then send an ack
// for sender socket.
// this is a private method and every peer can have one listen socket for recieved message.
func (p *ReliableUdpPeer) listen(address string) error {
	// variable for read from socket
	data := make([]byte, 2048)

	// resolve listeing address and get *net.UDPAddr
	// object of address
	addr, err := net.ResolveUDPAddr("udp", address)
	// if any error occured return it
	if err != nil {
		return err
	}

	// listen on address via udp socket
	ln, err := net.ListenUDP("udp", addr)
	// if any error occured return it
	if err != nil {
		return err
	}

	// call a goroutine for received messages
	go func() {
		// close socket after returing from function
		defer ln.Close()

		for {
			// wait for messages. when anyone received read
			// from socket and save it in data variable.
			// length is number of valid byte of data vaiable.
			// remoteAddr is *net.UDPAddr of message sender.
			length, remoteAddr, err := ln.ReadFromUDP(data)
			// if any error occured return it
			if err != nil {
				p.Log("%v", err)
			}

			// save received message. remove invalid bytes and convert to string.
			p.ReceivedBuffer = append(p.ReceivedBuffer, string(data[:length]))
			// log received message for debug
			p.Log("From %v: %s", remoteAddr, data[:length])

			// send ack message for message sender.
			_, err = ln.WriteToUDP([]byte("ack"), remoteAddr)
			// if any error occured return it
			if err != nil {
				p.Log("%v", err)
			}
		}
	}()
	p.Log("Listen on %v successfully", address)
	return nil
}

// get a connection socket of and destination address
// and send message to the address wait 1 second for
// ack message from receiver but don't retrasmit message.
// if any error occured in sending and acking return with error
// otherwise error is nil.
func (p *ReliableUdpPeer) Send(conn *net.UDPConn, destAddr *net.UDPAddr, msg string) error {
	// send message via udp socket
	_, err := conn.WriteToUDP([]byte(msg), destAddr)
	// if any error occured return it.
	if err != nil {
		return err
	}

	// create a variable for reading ack message
	data := make([]byte, 2048)
	// calculate deadline time for receiver to ack message
	deadline := time.Now().Add(time.Second)
	// set a deadline for ack message
	err = conn.SetReadDeadline(deadline)
	// if any error occured return it
	if err != nil {
		return err
	}

	// read from udp socket and save it in data variable
	_, remoteAddr, err := conn.ReadFromUDP(data)
	// if any error occured return it
	// include exporation of acking deadline
	if err != nil {
		return err
	}

	// log ack is success
	p.Log("%s acked", remoteAddr)
	// return without error
	return nil
}

// send message and any error occured (include not acked or ...) retrasmit it.
func (p *ReliableUdpPeer) ReliableSend(msg string, address string) error {
	// resolve destination address and get *net.UDPAddr of it
	destAddr, err := net.ResolveUDPAddr("udp", address)
	// if any error occured return it
	if err != nil {
		return err
	}

	// create a udp socket on a random port (second param is nil).
	conn, err := net.ListenUDP("udp", nil)
	// if any error occured return it
	if err != nil {
		return err
	}

	go func() {
		// close socket affter return. send is success and ack is recieved.
		defer conn.Close()
		// add to one to number of goroutine that wg wait for finish.
		p.wg.Add(1)

		// block goroutine if inflight message are exceeded from limit
		p.inFlightLimit <- nil

		// reliable sending loop
		for {
			// send message unreliable
			err = p.Send(conn, destAddr, msg)

			// check error
			if err != nil {
				p.Log("%v", err)
			} else {
				// if no error occurred decrease number of inflight message and goroutine for wait.
				<-p.inFlightLimit
				p.wg.Done()
				// breack from for loop and close routine
				break
			}
		}
	}()
	return nil
}

// wait until all message recieved and acked
func (p *ReliableUdpPeer) Wait() {
	p.wg.Wait()
}

// customize log for peer object
func (p *ReliableUdpPeer) Log(msg string, a ...interface{}) {
	p.logger.Printf("Peer %v: %v \n", p.name, fmt.Sprintf(msg, a...))
}
