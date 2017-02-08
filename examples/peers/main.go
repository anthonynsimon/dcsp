package main

import (
	"fmt"

	"flag"

	"log"

	"strings"

	"github.com/anthonynsimon/dcsp"
)

var (
	channelType *string
	addr        *string
)

func init() {
	channelType = flag.String("type", "", "the type of the peer, either 'sender' or 'receiver'")
	addr = flag.String("address", "", "the address to send to or receive from")
}

func main() {
	flag.Parse()

	if *addr == "" {
		log.Fatal("address cannot be empty")
	}

	switch *channelType {
	case "sender":
		sender(*addr)
	case "receiver":
		receiver(*addr)
	case "receiver-sender":
		receiversender(*addr)
	default:
		log.Fatal("unrecognized channel type:", *channelType)
	}
}

func sender(addr string) {
	sch := dcsp.NewSendChannel(dcsp.NewTCPTransport(addr))

	i := 0
	for {
		err := sch.Send([]byte(fmt.Sprintf("MESSAGE #%010d", i)))
		if err != nil {
			fmt.Println(err)
			return
		}
		i++
	}
}

func receiver(addr string) {
	rch := dcsp.NewReceiveChannel(dcsp.NewTCPTransport(addr))

	for {
		msg, err := rch.Receive()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("%s => RECEIVED\r\n", msg)
	}
}

func receiversender(addrs string) {
	addrParts := strings.Split(addrs, "--")
	if len(addrParts) != 2 || addrParts[0] == "" || addrParts[1] == "" {
		fmt.Println("receiver-sender requires two addresses")
		return
	}

	rch := dcsp.NewReceiveChannel(dcsp.NewTCPTransport(addrParts[0]))
	sch := dcsp.NewSendChannel(dcsp.NewTCPTransport(addrParts[1]))

	for {
		msg, err := rch.Receive()
		if err != nil {
			fmt.Println(err)
			return
		}
		err = sch.Send([]byte(fmt.Sprintf("%s => FORWARDED BY MIDDLEMAN", msg)))
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}
