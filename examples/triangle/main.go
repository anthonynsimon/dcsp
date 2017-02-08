package main

import (
	"fmt"

	"io"

	"log"

	"os"

	"bytes"

	"time"

	"github.com/anthonynsimon/dcsp"
)

func main() {
	go entitiy("ENTITY_1", "localhost:7260", "localhost:7261", true)
	go entitiy("ENTITY_2", "localhost:7262", "localhost:7260", false)
	go entitiy("ENTITY_3", "localhost:7261", "localhost:7262", false)
	// Block forever
	var done chan bool
	<-done
}

func entitiy(name, saddr, raddr string, sendFirst bool) {
	sch := dcsp.NewSendChannel(dcsp.NewTCPTransport(saddr))
	rch := dcsp.NewReceiveChannel(dcsp.NewTCPTransport(raddr))

	fileOut, err := os.Create(name + ".txt")
	if err != nil {
		log.Fatal(err)
	}
	defer fileOut.Close()

	outChan := make(chan []byte, 2048)
	go fileWriter(outChan, fileOut)

	if sendFirst {
		i := 0
		for {
			sch.Send([]byte(fmt.Sprintf("MESSAGE FROM %s #%010d", name, i)))
			time.Sleep(50 * time.Millisecond)
			msg := rch.Receive()
			fmt.Println(string(msg))
			if i > 40 {
				break
			}
			outChan <- msg
			i++
		}
	} else {
		i := 0
		for {
			time.Sleep(50 * time.Millisecond)
			msg := rch.Receive()
			fmt.Println(string(msg))
			outChan <- msg
			sch.Send([]byte(fmt.Sprintf("MESSAGE FROM %s #%010d", name, i)))
			i++
		}
	}

}

func fileWriter(ch chan []byte, w io.Writer) {
	for {
		select {
		case msg := <-ch:
			out := bytes.Join([][]byte{
				msg,
				[]byte("\r\n"),
			}, []byte{})
			_, err := w.Write(out)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
