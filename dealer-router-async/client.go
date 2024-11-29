// Asynchronous client-server

package main

import (
	"fmt"
	"os"
	"time"

	zmq "github.com/pebbe/zmq4"
)

type ClientTask struct {
    id string
}

func (c *ClientTask) run() {
    context, _ := zmq.NewContext()
    socket, _ := context.NewSocket(zmq.DEALER)
    defer socket.Close()
    defer context.Term()

    socket.SetIdentity(c.id)
    socket.Connect("tcp://localhost:5570")
    fmt.Printf("Client %s started\n", c.id)

    poller := zmq.NewPoller()
    poller.Add(socket, zmq.POLLIN)

    reqs := 0
    for {
        reqs++
        fmt.Printf("Req #%d sent..\n", reqs)
        socket.Send(fmt.Sprintf("request #%d", reqs), 0)

        time.Sleep(time.Second)
        
        sockets, _ := poller.Poll(time.Second)
        if len(sockets) > 0 {
            msg, _ := socket.RecvMessageBytes(0)
            fmt.Printf("%s received: %s\n", c.id, msg[0])
        }
    }
}

func main() {
    client := &ClientTask{id: os.Args[1]}
    client.run()
}