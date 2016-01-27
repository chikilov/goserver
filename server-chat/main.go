package main

import (
	"goserver/common"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

const (
	CONN_HOST = "0.0.0.0"
	CONN_PORT = "3333"
	CONN_TYPE = "tcp"
)

type netCon struct {
	connection     net.Conn
	username       string
	isHandShakeing bool
}

var (
	netCons []netCon
)

func main() {
	// Listen for incoming connections.
	l, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()
	fmt.Println("Listening on " + CONN_HOST + ":" + CONN_PORT)
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		var newCon []netCon
		var tmpCon netCon
		tmpCon.connection = conn
		tmpCon.isHandShakeing = false
		newCon = append(newCon, tmpCon)
		// Save connection
		netCons = append(netCons, newCon...)
		fmt.Printf("--------\n")
		// Handle connections in a new goroutine.
		go handleRequest(newCon[0])
	}
}

// Handles incoming requests.
func handleRequest(newCon netCon) {
	for {
		msg, err := common.ReadMsg(newCon.connection)
		if err != nil {
			if err == io.EOF {
				// Close the connection when you're done with it.
				removeConn(newCon.connection)
				newCon.connection.Close()
				return
			}
			log.Println(err)
			return
		}
		if newCon.isHandShakeing {
			fmt.Printf("Message Received: %s\n", msg)
			broadcast(newCon, msg)
		} else {
			var y map[string]interface{}
			json.Unmarshal([]byte(msg), &y)
			for n, v := range y {
				if n == "username" {
					for i := range netCons {
						if netCons[i] == newCon {
							netCons[i].isHandShakeing = true
							netCons[i].username = v.(string)
							newCon = netCons[i]
						}
					}
				}
			}
			fmt.Printf("Handshake success: %s\n", newCon.username)
			broadcast(newCon, newCon.username+" has join now")
		}
	}
}

func removeConn(conn net.Conn) {
	var i int
	for i = range netCons {
		if netCons[i].connection == conn {
			break
		}
	}
	fmt.Println(i)
	netCons = append(netCons[:i], netCons[i+1:]...)
}

func broadcast(newCon netCon, msg string) {
	for i := range netCons {
		if netCons[i].isHandShakeing {
			err := common.WriteMsg(netCons[i].connection, msg)
			if err != nil {
				log.Println(err)
			}
		}
	}
}
