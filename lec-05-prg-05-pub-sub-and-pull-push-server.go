package main

import (
	"fmt"

	zmq "github.com/pebbe/zmq4"
)

func main() {
	// context and socket
	context, _ := zmq.NewContext()
	defer context.Term()

	publisher, _ := context.NewSocket(zmq.PUB)
	defer publisher.Close()
	publisher.Bind("tcp://localhost:5557")

	collector, _ := context.NewSocket(zmq.PULL)
	defer collector.Close()
	collector.Bind("tcp://localhost:5558")

	for {
		message, _ := collector.Recv(0)
		fmt.Println("I: publishing update ", message)
		publisher.Send(message, 0)
	}
}