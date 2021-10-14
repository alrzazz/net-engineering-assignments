package main

import (
	"fmt"
	"log"
	"os"

	rup "github.com/alrzazz/net-engineering-assignments/01-reliable-udp-peer"
)

func main() {

	logger := log.New(os.Stdout, "socket-log ", log.Lmicroseconds)

	pa := rup.NewReliableUdpPeer(logger, "Alice")
	pb := rup.NewReliableUdpPeer(logger, "Bob")

	pa.Listen("127.0.0.1:6666")
	pb.Listen("127.0.0.1:5555")

	pa.ReliableSend("Alice: Hi Bob.", "127.0.0.1:5555")
	pa.ReliableSend("Alice: Are you ok?", "127.0.0.1:5555")
	pb.ReliableSend("Bob: Hello Alice.", "127.0.0.1:6666")

	pa.Wait()
	pb.Wait()

	fmt.Printf("Alice Recieved Buffer: %s\n", pa.ReceivedBuffer)
	fmt.Printf("Bob Recieved Buffer: %s\n", pb.ReceivedBuffer)

	slice := [...]int{10, 20, 30, 40, 50}

	slice2 := slice[1:3]

	fmt.Printf("%T %T", slice, slice2)

}
