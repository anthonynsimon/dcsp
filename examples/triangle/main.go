package main

import (
	"fmt"

	"io"

	"log"

	"os"

	"time"

	"github.com/anthonynsimon/dcsp"
)

func main() {
	// Simulate network entities
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

	if sendFirst {
		fileOut, err := os.Create(name + ".txt")
		if err != nil {
			log.Fatal(err)
		}
		defer fileOut.Close()

		outChan := make(chan string, 2048)
		go fileWriter(outChan, fileOut)

		i := 0
		for {
			sch.Send([]byte(fmt.Sprintf("#%010d sent from %s", i, name)))
			msg := rch.Receive()
			outChan <- fmt.Sprintf("%s => received by %s\r\n", msg, name)
			i++
		}
	} else {
		for {
			time.Sleep(50 * time.Millisecond)
			msg := rch.Receive()
			sch.Send([]byte(fmt.Sprintf("%s => %s", msg, name)))
		}
	}

}

func fileWriter(ch chan string, w io.Writer) {
	for {
		select {
		case msg := <-ch:
			_, err := w.Write([]byte(msg))
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
