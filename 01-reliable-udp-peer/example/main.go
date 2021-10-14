package main

import (
	"fmt"
	"log"
	"os"

	rup "github.com/alrzazz/net-engineering-assignments/01-reliable-udp-peer"
)

func main() {

	// create a *log.Logger with show time of every log for debuging in go routine
	logger := log.New(os.Stdout, "socket-log ", log.Lmicroseconds)

	// create ReliableUdpPeer and listen
	pa, err := rup.NewReliableUdpPeer(logger, "Alice", "127.0.0.1:6666", 2)
	if err != nil {
		panic(err)
	}

	// create ReliableUdpPeer and listen
	pb, err := rup.NewReliableUdpPeer(logger, "Bob", "127.0.0.1:5555", 2)
	if err != nil {
		panic(err)
	}

	// send message for Bob from Alice peer
	err = pa.ReliableSend("Alice: Hi Bob.", "127.0.0.1:5555")
	if err != nil {
		panic(err)
	}

	// send message for Bob from Alice peer
	pa.ReliableSend("Alice: Are you ok?", "127.0.0.1:5555")
	if err != nil {
		panic(err)
	}

	// send message for Bob from Alice peer
	pa.ReliableSend("Alice: Are you ok2?", "127.0.0.1:5555") // this message wait for previous message to recieve
	if err != nil {
		panic(err)
	}

	// send message for Alice from Bob peer
	pb.ReliableSend("Bob: Hello Alice.", "127.0.0.1:6666")
	if err != nil {
		panic(err)
	}

	// wait for all inflight messages
	pa.Wait()
	pb.Wait()

	fmt.Printf("Alice Recieved Buffer: %s\n", pa.ReceivedBuffer)
	fmt.Printf("Bob Recieved Buffer: %s\n", pb.ReceivedBuffer)

}

// result:
// socket-log 17:34:23.435496 Peer Alice: Listen on 127.0.0.1:6666 successfully
// socket-log 17:34:23.436544 Peer Bob: Listen on 127.0.0.1:5555 successfully
// socket-log 17:34:23.437869 Peer Alice: From 127.0.0.1:56673: Bob: Hello Alice.
// socket-log 17:34:23.437869 Peer Bob: From 127.0.0.1:56670: Alice: Hi Bob.
// socket-log 17:34:23.438439 Peer Bob: 127.0.0.1:6666 acked
// socket-log 17:34:23.439032 Peer Alice: 127.0.0.1:5555 acked
// socket-log 17:34:23.439544 Peer Bob: From 127.0.0.1:56671: Alice: Are you ok?
// socket-log 17:34:23.440819 Peer Alice: 127.0.0.1:5555 acked
// socket-log 17:34:23.440819 Peer Bob: From 127.0.0.1:56672: Alice: Are you ok2?
// socket-log 17:34:23.442230 Peer Alice: 127.0.0.1:5555 acked
// Alice Recieved Buffer: [Bob: Hello Alice.]
// Bob Recieved Buffer: [Alice: Hi Bob. Alice: Are you ok? Alice: Are you ok2?]
