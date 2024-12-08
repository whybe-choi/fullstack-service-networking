//
// Hello World Zeromq Server
// Binds REP socket to tcp://*:5555
// Expects "Hello" from client, replies with "World"

package main

import (
	"fmt"
	"time"

	zmq "github.com/pebbe/zmq4"
)

func main() {
	context, _ := zmq.NewContext()
	defer context.Term()

	socket, _ := context.NewSocket(zmq.REP)
	defer socket.Close()

	socket.Bind("tcp://*:5555")

	for {
		// Wait for next requirest from client
		message, _ := socket.Recv(0)
		fmt.Printf("Received request: %s\n", message)

		// Do some 'work
		time.Sleep(1 * time.Second)

		// Send reply back to client
		socket.Send("World", 0)
	}
}