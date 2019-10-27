// This main file shows an example of UDP broadcast and receive.
//
// It should be run like
//
//    docker run --rm -ti -v "$(pwd):/cwd" golang:1.13 go run /cwd/peer/peer.go
//
// When running just one instance of this program, you should see the peer receiving its own broadcasts.
// When running multiple copies of this program, you should see each peer receiving each other peers messages.

package main

import (
	"log"
	"math/rand"
	"net"
	"strconv"
	"time"
)

// We sent to, and listen on, the same port.
const commonPort = 2002

func main() {
	// Construct a random peer name for myself which I'll broadcast over the network
	rand.Seed(time.Now().UnixNano())
	name := "peer-" + strconv.Itoa(rand.Int())
	log.Printf("I am %s", name)

	// Resolve broadcast address into a address object for udp network
	log.Printf("Resolving broadcast address..")
	bCastAddr, err := net.ResolveUDPAddr("udp", "255.255.255.255:" + strconv.Itoa(commonPort))
	if err != nil {
		log.Fatalf("Error: Failed to resolve broadcast address: %s", err.Error())
	}
	log.Printf("Resolved %s.", bCastAddr)

	// Background thread
	go func() {
		// "Connect" to udp address (assigns local source port)
		log.Printf("Dialing broadcast address..")
		conn, err := net.DialUDP("udp", nil, bCastAddr)
		if err != nil {
			log.Fatalf("Error: Failed to dial: %s", err.Error())
		}
		// Start sending packets off into the ether every second
		for {
			time.Sleep(time.Second)
			log.Printf("Broadcasting..")
			if _, err := conn.Write([]byte(name)); err != nil {
				log.Printf("Error: %s", err.Error())
			}
			// We could choose to listen here on our source port and wait for direct replies
			// but since we want to listen on the broadcast port ourselves and don't actually expect
			// any direct replies, we wont.
		}
	}()

	// Resolve the broadcast listen address on our own interface
	log.Printf("Resolving listen address..")
	listenAddr, err := net.ResolveUDPAddr("udp", ":" + strconv.Itoa(commonPort))
	if err != nil {
		log.Fatalf("Error: Failed to resolve listen address: %s", err.Error())
	}
	log.Printf("Resolved %s.", listenAddr)

	// Now begin listing on the port, we'll
	log.Printf("Beginning to listen..")
	conn, err := net.ListenUDP("udp", listenAddr)
	if err != nil {
		log.Fatalf("Error: Failed to listen: %s", err.Error())
	}
	defer conn.Close()

	// Buffer for reading received datagrams into, this is really up to the protocol and since in this case
	// we know we're only sending short strings, we'll just use 1K. Remember, general UDP datagrams can be fragmented
	// into multiple IP packets so this can be as high as 65,536 bytes since the IP "Total Length" field is 2 bytes and
	// uses a "Fragment Offset" field of 13 bits (2**13 – 1) × 8 = 65,528 bytes.
	// If the buffer is too small, the data will be discarded during copy from the OS.
	buf := make([]byte, 1024)
	for {
		// Read some bytes from our listened connection and get the source addr. If we wanted to, we could write
		// bytes directly back to that source address (if it was waiting to read on that). This is generally
		// what you want to do to initiate a direct communication with the given host in order to exchange specific
		// information.
		n, addr, err := conn.ReadFromUDP(buf)
		log.Printf("Received '%s' from %s", string(buf[0:n]), addr)
		if err != nil {
			log.Printf("Error: %s", err.Error())
		}
	}

}
