package main

import (
	"fmt"
	"log"
	"net"
)

// type Message struct {
// 		message string
// }

func main() {
	writeEvent()
}

func open(addr string) net.Conn {
	conn, err := net.Dial("tcp", "localhost:7777")
	if err != nil {
		fmt.Println("Dialing "+addr+" failed:", err)
	}

	log.Println("Connected to tcp listener")

	return conn
}


func writeEvent() {
	requestUrl := "http://localhost:7777/"

	jsonStr := []byte(`{"everyone":false,"spawnMissile":{"nid":"5577006791947779419","owner":"5577006791947779418","pos":{"x":6.7987933,"y":12.237321},"momentum":{"x":-14.322695,"y":8.191533},"rot":2.6179938}}`)

	c := open(requestUrl)
	defer c.Close()

	// go func() {

	log.Println("Sending to listener: ", jsonStr)
	_, err := c.Write([]byte(jsonStr))
	if err != nil {
		log.Println("Error writing data to tcp connection: ", err)
	}
	// }()
}
