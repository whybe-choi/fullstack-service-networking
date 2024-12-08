package main

import (
	"fmt"

	zmq "github.com/pebbe/zmq4"
)

func main() {
    // context and sockets
    ctx, _ := zmq.NewContext()
	defer ctx.Term()

    publisher, _ := ctx.NewSocket(zmq.PUB)
	defer publisher.Close()
    publisher.Bind("tcp://*:5557")
	
    collector, _ := ctx.NewSocket(zmq.PULL)
	defer collector.Close()
    collector.Bind("tcp://*:5558")

    for {
        message, _ := collector.Recv(0)
        fmt.Println("server: publishing update => ", message)
        publisher.Send(message, 0)
    }
}