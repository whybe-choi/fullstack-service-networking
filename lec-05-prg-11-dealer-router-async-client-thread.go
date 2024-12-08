package main

import (
	"fmt"
	"os"
	"time"

	zmq "github.com/pebbe/zmq4"
)

type ClientTask struct {
	id       string
	identity string
	context  *zmq.Context
	socket   *zmq.Socket
	poller   *zmq.Poller
}

func NewClientTask(id string) *ClientTask {
	return &ClientTask{id: id}
}

func (c *ClientTask) recvHandler() {
	for {
		sockets, err := c.poller.Poll(1000 * time.Millisecond)
		if err != nil {
			continue
		}

		for _, s := range sockets {
			if s.Socket == c.socket {
				msg, err := c.socket.Recv(0)
				if err == nil {
					fmt.Printf("%s received: %s\n", c.identity, msg)
				}
			}
		}
	}
}

func (c *ClientTask) run() {
	c.context, _ = zmq.NewContext()
	c.socket, _ = c.context.NewSocket(zmq.DEALER)
	c.identity = c.id
	c.socket.SetIdentity(c.identity)
	c.socket.Connect("tcp://localhost:5570")

	fmt.Printf("Client %s started\n", c.identity)

	c.poller = zmq.NewPoller()
	c.poller.Add(c.socket, zmq.POLLIN)

	go c.recvHandler()

	reqs := 0
	for {
		reqs++
		fmt.Printf("Req #%d sent..\n", reqs)
		c.socket.Send(fmt.Sprintf("request #%d", reqs), 0)
		time.Sleep(1 * time.Second)
	}

	c.socket.Close()
	c.context.Term()
}

func main() {
	clientID := os.Args[1]
	client := NewClientTask(clientID)
	client.run()
}