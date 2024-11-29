//  Asynchronous client-server

package main

import (
	"fmt"
	"os"
	"strconv"

	zmq "github.com/pebbe/zmq4"
)

type ServerWorker struct {
    context *zmq.Context
    id      int
}

func (w *ServerWorker) run() {
    worker, _ := w.context.NewSocket(zmq.DEALER)
    defer worker.Close()
    worker.Connect("inproc://backend")
    
    fmt.Printf("Worker#%d started\n", w.id)
    
    for {
        parts, _ := worker.RecvMessageBytes(0)
        ident, msg := parts[0], parts[1]
        fmt.Printf("Worker#%d received %s from %s\n", w.id, msg, ident)
        worker.SendMessage(parts)
    }
}

func server_task(numWorkers int) {
    context, _ := zmq.NewContext()
    defer context.Term()

    frontend, _ := context.NewSocket(zmq.ROUTER)
    defer frontend.Close()
    frontend.Bind("tcp://*:5570")

    backend, _ := context.NewSocket(zmq.DEALER)
    defer backend.Close()
    backend.Bind("inproc://backend")

    for i := 0; i < numWorkers; i++ {
        worker := &ServerWorker{
            context: context,
            id:      i,
        }
        go worker.run()
    }

    zmq.Proxy(frontend, backend, nil)
}

func main() {
    numWorkers := 5
    if len(os.Args) > 1 {
        if n, err := strconv.Atoi(os.Args[1]); err == nil {
            numWorkers = n
        }
    }
    
    server_task(numWorkers)
}