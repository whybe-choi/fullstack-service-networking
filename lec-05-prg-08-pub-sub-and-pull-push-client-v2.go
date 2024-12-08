package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	zmq "github.com/pebbe/zmq4"
)

func main() {
    // Prepare our context and subscriber
    ctx, _ := zmq.NewContext()
	defer ctx.Term()

    subscriber, _ := ctx.NewSocket(zmq.SUB)
	defer subscriber.Close()
    subscriber.Connect("tcp://localhost:5557")
	subscriber.SetSubscribe("") // Subscribe to all messages

    publisher, _ := ctx.NewSocket(zmq.PUSH)
	defer publisher.Close()
    publisher.Connect("tcp://localhost:5558")

    clientID := os.Args[1]
    rand.New(rand.NewSource(time.Now().UnixNano()))

	poller := zmq.NewPoller()
	poller.Add(subscriber, zmq.POLLIN)

	for {
		sockets, _ := poller.Poll(1 * time.Second)
		if len(sockets) > 0 {
			message, _ := subscriber.Recv(0)
			fmt.Printf("%s: receive status => %s\n", clientID, message)
		} else {
			randNum := rand.Intn(100) + 1
			if randNum < 10 {
				time.Sleep(1 * time.Second)
				msg := fmt.Sprintf("(%s:ON)", clientID)
				publisher.Send(msg, 0)
				fmt.Printf("%s: send status - activated\n", clientID)
			} else if randNum > 90 {
				time.Sleep(1 * time.Second)
				msg := fmt.Sprintf("(%s:OFF)", clientID)
				publisher.Send(msg, 0)
				fmt.Printf("%s: send status - deactivated\n", clientID)
			}
		}
	}
}