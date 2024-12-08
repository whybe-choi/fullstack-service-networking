package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	zmq "github.com/pebbe/zmq4"
	// 실제 IP 획득을 위해 net, os/exec 등을 사용 가능하지만 여기서는 단순화
)

// Python 코드와 같은 로직 구현

func searchNameserver(ipMask, localIPAddr string, portNameserver int) *string {
	context, _ := zmq.NewContext()
	req, _ := context.NewSocket(zmq.SUB)
	defer req.Close()

	// 모든 후보 IP에 대해 접속 시도
	for last := 1; last < 255; last++ {
		targetIPAddr := fmt.Sprintf("tcp://%s.%d:%d", ipMask, last, portNameserver)
		// Python 코드 로직상 이 조건은 항상 참
		req.Connect(targetIPAddr)
		req.SetRcvtimeo(2000) // 2초 타임아웃
		req.SetSubscribe("NAMESERVER")
	}

	// 한번만 recv 시도
	res, err := req.RecvMessage(0)
	if err != nil {
		return nil
	}

	// "NAMESERVER:xxx.xxx.xxx.xxx" 형태인지 검사
	resList := strings.Split(res[0], ":")
	if len(resList) > 1 && resList[0] == "NAMESERVER" {
		addr := resList[1]
		return &addr
	}
	return nil
}

func beaconNameserver(localIPAddr string, portNameserver int) {
	context, _ := zmq.NewContext()
	socket, _ := context.NewSocket(zmq.PUB)
	defer socket.Close()
	socket.Bind(fmt.Sprintf("tcp://%s:%d", localIPAddr, portNameserver))
	fmt.Printf("local p2p name server bind to tcp://%s:%d.\n", localIPAddr, portNameserver)

	for {
		time.Sleep(1 * time.Second)
		msg := fmt.Sprintf("NAMESERVER:%s", localIPAddr)
		socket.Send(msg, 0)
	}
}

func userManagerNameserver(localIPAddr string, portSubscribe int) {
	userDB := [][]string{}
	context, _ := zmq.NewContext()
	socket, _ := context.NewSocket(zmq.REP)
	defer socket.Close()

	socket.Bind(fmt.Sprintf("tcp://%s:%d", localIPAddr, portSubscribe))
	fmt.Printf("local p2p db server activated at tcp://%s:%d.\n", localIPAddr, portSubscribe)

	for {
		userReqStr, err := socket.Recv(0)
		if err != nil {
			continue
		}
		userReq := strings.Split(userReqStr, ":")
		if len(userReq) == 2 {
			userDB = append(userDB, userReq)
			fmt.Printf("user registration '%s' from '%s'.\n", userReq[1], userReq[0])
			socket.Send("ok", 0)
		} else {
			socket.Send("fail", 0)
		}
	}
}

func relayServerNameserver(localIPAddr string, portChatPublisher, portChatCollector int) {
	context, _ := zmq.NewContext()
	publisher, _ := context.NewSocket(zmq.PUB)
	collector, _ := context.NewSocket(zmq.PULL)

	publisher.Bind(fmt.Sprintf("tcp://%s:%d", localIPAddr, portChatPublisher))
	collector.Bind(fmt.Sprintf("tcp://%s:%d", localIPAddr, portChatCollector))
	fmt.Printf("local p2p relay server activated at tcp://%s:%d & %d.\n", localIPAddr, portChatPublisher, portChatCollector)

	for {
		message, err := collector.Recv(0)
		if err != nil {
			continue
		}
		fmt.Printf("p2p-relay:<==> %s\n", message)
		publisher.Send(fmt.Sprintf("RELAY:%s", message), 0)
	}
}

func getLocalIP() string {
	// 원본 Python 코드와 동일한 로직으로 IP 얻으려면 추가 구현 필요
	// 여기서는 임시로 로컬 IP 반환
	return "127.0.0.1"
}

func mainLogic(argv []string) {
	var ipAddrP2pServer string
	portNameserver := 9001
	portChatPublisher := 9002
	portChatCollector := 9003
	portSubscribe := 9004

	userName := argv[1]
	ipAddr := getLocalIP()
	ipMask := ipAddr[:strings.LastIndex(ipAddr, ".")]

	fmt.Println("searching for p2p server.")

	nameServerIPAddr := searchNameserver(ipMask, ipAddr, portNameserver)
	if nameServerIPAddr == nil {
		ipAddrP2pServer = ipAddr
		fmt.Println("p2p server is not found, and p2p server mode is activated.")
		beaconThread := func() { beaconNameserver(ipAddr, portNameserver) }
		dbThread := func() { userManagerNameserver(ipAddr, portSubscribe) }
		relayThread := func() { relayServerNameserver(ipAddr, portChatPublisher, portChatCollector) }

		go beaconThread()
		fmt.Println("p2p beacon server is activated.")
		go dbThread()
		fmt.Println("p2p subscriber database server is activated.")
		go relayThread()
		fmt.Println("p2p message relay server is activated.")
	} else {
		ipAddrP2pServer = *nameServerIPAddr
		fmt.Printf("p2p server found at %s, and p2p client mode is activated.\n", ipAddrP2pServer)
	}

	fmt.Println("starting user registration procedure.")

	dbClientContext, _ := zmq.NewContext()
	dbClientSocket, _ := dbClientContext.NewSocket(zmq.REQ)
	dbClientSocket.Connect(fmt.Sprintf("tcp://%s:%d", ipAddrP2pServer, portSubscribe))
	dbClientSocket.Send(fmt.Sprintf("%s:%s", ipAddr, userName), 0)
	reply, _ := dbClientSocket.Recv(0)
	if reply == "ok" {
		fmt.Println("user registration to p2p server completed.")
	} else {
		fmt.Println("user registration to p2p server failed.")
	}

	fmt.Println("starting message transfer procedure.")

	relayClient, _ := zmq.NewContext()
	p2pRx, _ := relayClient.NewSocket(zmq.SUB)
	p2pTx, _ := relayClient.NewSocket(zmq.PUSH)

	p2pRx.SetSubscribe("RELAY")
	p2pRx.Connect(fmt.Sprintf("tcp://%s:%d", ipAddrP2pServer, portChatPublisher))
	p2pTx.Connect(fmt.Sprintf("tcp://%s:%d", ipAddrP2pServer, portChatCollector))

	fmt.Println("starting autonomous message transmit and receive scenario.")

	poller := zmq.NewPoller()
	poller.Add(p2pRx, zmq.POLLIN)

	for {
		randVal := rand.Intn(100) + 1
		if randVal < 10 {
			time.Sleep(3 * time.Second)
			msg := fmt.Sprintf("(%s,%s:ON)", userName, ipAddr)
			p2pTx.Send(msg, 0)
			fmt.Printf("p2p-send::==>> %s\n", msg)
		} else if randVal > 90 {
			time.Sleep(3 * time.Second)
			msg := fmt.Sprintf("(%s,%s:OFF)", userName, ipAddr)
			p2pTx.Send(msg, 0)
			fmt.Printf("p2p-send::==>> %s\n", msg)
		}

		sockets, err := poller.Poll(100 * time.Millisecond)
		if err != nil {
			continue
		}
		for _, s := range sockets {
			if s.Socket == p2pRx {
				message, _ := p2pRx.Recv(0)
				parts := strings.Split(message, ":")
				if len(parts) > 2 {
					fmt.Printf("p2p-recv::<<== %s:%s\n", parts[1], parts[2])
				}
			}
		}
	}
}

func main() {
	if len(os.Args) == 1 {
		fmt.Println("usage is 'python dechat.py _user-name_'.")
	} else {
		fmt.Println("starting p2p chatting program.")
		mainLogic(os.Args)
	}
}