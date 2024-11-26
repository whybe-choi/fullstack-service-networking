//
// Weather update client
// Connects SUB sockets to tcp://localhost:5556
// Collects weather updates and finds avg temp in zipcode
//

package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	zmq "github.com/pebbe/zmq4"
)

func main() {
	// Socket to talk to server
	context, _ := zmq.NewContext()
	defer context.Term()

	socket, _ := context.NewSocket(zmq.SUB)
	defer socket.Close()

	fmt.Println("Collecting updates from weather server...")
	socket.Connect("tcp://localhost:5556")

	// Subscribe to zipcode, default is NYC, 10001
	zip_filter := "10001"

	if len(os.Args) > 1 {
		zip_filter = string(os.Args[1])
	}

	socket.SetSubscribe(zip_filter)

	// 변수 선언
	total_temp := 0
	update_nbr := 0

	// Process 20 updates
	for ; update_nbr < 20; update_nbr++ {
		datapt, _ := socket.Recv(0)
		data := strings.Split(string(datapt), " ")
		temperature := data[1]
		temp, _ := strconv.Atoi(temperature)
		total_temp += temp

		fmt.Printf("Receive temperature for zipcode '%s' was %s F\n", 
			zip_filter, temperature)
	}

	fmt.Printf("Average temperature for zipcode '%s' was %d F\n", 
		zip_filter, total_temp/update_nbr)
}