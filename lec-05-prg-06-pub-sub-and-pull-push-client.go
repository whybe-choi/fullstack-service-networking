package main

import (
	"fmt"
	"math/rand"
	"time"

	zmq "github.com/pebbe/zmq4"
)

func main() {
	// Prepare our context and subscriber
	context, _ := zmq.NewContext()
	defer context.Term()

	subscriber, _ := context.NewSocket(zmq.SUB)
	defer subscriber.Close()
	subscriber.Connect("tcp://localhost:5557")
	subscriber.SetSubscribe("")

	publisher, _ := context.NewSocket(zmq.PUSH)
	defer publisher.Close()
	publisher.Connect("tcp://localhost:5558")

	rand.New(rand.NewSource(time.Now().UnixNano()))

	poller := zmq.NewPoller()
	poller.Add(subscriber, zmq.POLLIN)

	for {
		sockets, _ := poller.Poll(1 * time.Second)
		if len(sockets) > 0 {
			message, _ := subscriber.Recv(0)
			println("I: received message", message)
		} else {
			randNum := rand.Intn(100) + 1
			if randNum < 10 {
				publisher.Send(fmt.Sprintf("%d", randNum), 0)
				println("I: sending message", randNum)
			}
		}
	}
}
