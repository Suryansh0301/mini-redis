package main

import (
	"bufio"
	"fmt"
	parser "github.com/suryansh0301/mini-redis/internal/core/protocol/resp"
	"log/slog"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	slog.SetLogLoggerLevel(-4)
	slog.Debug("Listening on port 8080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {

	defer conn.Close()
	reader := bufio.NewReader(conn)
	//writer := bufio.NewWriter(conn)
	readBuffer := make([]byte, 1024)
	//writeBuffer := make([]byte, 1024) //placeholder
	for {
		n, err := reader.Read(readBuffer)
		if err != nil {
			return
		}
		buffer := readBuffer[:n]
		fmt.Println(string(buffer))
		response := parser.Parse(buffer)
		fmt.Printf("%#+v", response)
		fmt.Printf("%#+v", response.Resp)
		value, err := parser.Decoder(response)
		fmt.Printf("%#+v,%+v", value, err)
		//execution -> for example the response is set into write buffer (just to figure out things)
		//_, err = writer.Write(writeBuffer)
		//if err != nil {
		//	return
		//}

		//err = writer.Flush()
		//if err != nil {
		//	return
		//}
	}
}
