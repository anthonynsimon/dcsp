package main

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/anthonynsimon/dcsp"
)

func main() {
	// Simulate network entities
	go sender("localhost:7261", "value A")
	go sender("localhost:7262", "value B")
	go sender("localhost:7263", "value C")
	go receiver("localhost:7261", "localhost:7262", "localhost:7263")
	dcsp.SetLogLevel(logrus.DebugLevel)
	// Block forever
	var done chan bool
	<-done
}

func sender(addr, msg string) {
	sch := dcsp.NewSendChannel(dcsp.NewTCPTransport(addr))

	i := 0
	for {
		err := sch.Send([]byte(msg))
		if err != nil {
			fmt.Println(err)
			return
		}
		i++
	}
}

func receiver(addr0, addr1, addr2 string) {
	r0 := dcsp.NewReceiveChannel(dcsp.NewTCPTransport(addr0))
	r1 := dcsp.NewReceiveChannel(dcsp.NewTCPTransport(addr1))
	r2 := dcsp.NewReceiveChannel(dcsp.NewTCPTransport(addr2))

	for {
		dcsp.Select(
			dcsp.NewSelector(r0, func() {
				msg, err := r0.Receive()
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Printf("%s => RECEIVED\r\n", msg)
			}),
			dcsp.NewSelector(r1, func() {
				msg, err := r1.Receive()
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Printf("%s => RECEIVED\r\n", msg)
			}),
			dcsp.NewSelector(r2, func() {
				msg, err := r2.Receive()
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Printf("%s => RECEIVED\r\n", msg)
			}),
		)
	}
}
