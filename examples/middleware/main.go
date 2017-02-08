package main

import (
	"fmt"

	"bytes"

	"github.com/Sirupsen/logrus"
	"github.com/anthonynsimon/dcsp"
)

func main() {
	// Simulate network entities
	go sender("localhost:7260")
	go receiver("localhost:7260")
	dcsp.SetLogLevel(logrus.ErrorLevel)
	// Block forever
	var done chan bool
	<-done
}

func sender(addr string) {
	var msgList [][]byte

	sch := dcsp.NewSendChannel(dcsp.NewTCPTransport(addr),
		func(m []byte) []byte {
			return bytes.Join(
				[][]byte{
					m,
					[]byte("=> added by sender middleware before sending"),
					[]byte("=> you can concatenate multiple byte streams"),
				},
				[]byte(" "),
			)
		},
		func(m []byte) []byte {
			// Collect the messages as they pass
			msgList = append(msgList, m)
			return m
		},
		//func(m []byte) []byte {
		//	// Or even replace them
		//	return []byte("message hijacked")
		//},
	)

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
	rch := dcsp.NewReceiveChannel(dcsp.NewTCPTransport(addr),
		func(m []byte) []byte {
			return bytes.Join(
				[][]byte{
					m,
					[]byte("=> added by receiver middleware before receiving"),
				},
				[]byte(" "),
			)
		},
	)

	for {
		msg, err := rch.Receive()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("%s => RECEIVED\r\n", msg)
	}
}
