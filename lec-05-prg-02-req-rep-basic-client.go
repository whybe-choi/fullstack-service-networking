//
// Hello World Zeromq Cleint
// Connects REQ socket to tcp://localhost:5555
// Sends "Hello" to server, expects "World" back
//

package main

import (
	"fmt"

	zmq "github.com/pebbe/zmq4"
)

func main() {
	context, _ := zmq.NewContext()
	defer context.Term()

	socket, _ := context.NewSocket(zmq.REQ)
	defer socket.Close()

	// Socket to talk to server
	fmt.Println("Connecting to the hello world server...")
	socket.Connect("tcp://localhost:5555")

	// Do 10 requests, waiting each time for a response
	for request := 0; request < 10; request++ {
		fmt.Printf("Sending request %d...\n", request)
		socket.Send("Hello", 0)

		// Get the reply
		message, _ := socket.Recv(0)
		fmt.Printf("Received reply %d [ %s ]\n", request, message)
	}
}