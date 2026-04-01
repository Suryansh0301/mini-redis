package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"

	"github.com/suryansh0301/mini-redis/internal/core/common"
	"github.com/suryansh0301/mini-redis/internal/core/datastore"
	parser "github.com/suryansh0301/mini-redis/internal/core/protocol/resp"
)

type client struct {
	Reader       *bufio.Reader
	Writer       *bufio.Writer
	ResponseChan chan common.RespValue
	ParserBuffer []byte
	ReadBuffer   []byte
	WriteBuffer  []byte
}

func newClient(connection net.Conn) *client {
	reader := bufio.NewReader(connection)
	writer := bufio.NewWriter(connection)

	return &client{
		Reader:       reader,
		Writer:       writer,
		ResponseChan: make(chan common.RespValue, 256),
		ParserBuffer: make([]byte, 0, 4096),
		ReadBuffer:   make([]byte, 1024, 4096),
		WriteBuffer:  make([]byte, 1024, 4096),
	}
}

func (c *client) read() (int, error) {
	n, err := c.Reader.Read(c.ReadBuffer)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (c *client) appendParseBuffer(n int) {
	c.ParserBuffer = append(c.ParserBuffer, c.ReadBuffer[:n]...)
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	slog.SetLogLoggerLevel(-4)
	slog.Debug("Listening on port 6379")

	exec := datastore.NewExecutor()

	go func() {
		for value := range exec.ExecutorChan {
			response := exec.Execute(value.Command)
			value.ResponseChan <- response
		}
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}

		client := newClient(conn)
		go client.handleConnection(conn, exec)
	}
}

func (c *client) handleConnection(conn net.Conn, exec *datastore.Executor) {

	defer conn.Close()

	go func() {
		for resp := range c.ResponseChan {
			byteResp := parser.Encoder(resp)
			c.WriteBuffer = append(c.WriteBuffer, byteResp...)
			n, err := c.Writer.Write(c.WriteBuffer)
			if err != nil {
				return
			}
			if n > 0 {
				err = c.Writer.Flush()
				if err != nil {
					return
				}
			}
		}
	}()

	for {
		n, err := c.read()
		if err != nil {
			return
		}

		c.appendParseBuffer(n)

		for len(c.ParserBuffer) > 0 {
			response := parser.Parse(c.ParserBuffer)
			if response.Error() != nil {
				// we receive an error response
				if err = c.handleError(response.Error().Error()); err != nil {

				}
				return
			}

			if response.BytesConsumed() == 0 {
				// we need more data hence we break and wait for the next read
				break
			}

			c.ParserBuffer = c.ParserBuffer[response.BytesConsumed():]
			value, err := parser.Decoder(response)
			if err != nil {
				if err = c.handleError(err.Error()); err != nil {

				}
				return
			}

			exec.ExecutorChan <- datastore.Value{
				Command:      value,
				ResponseChan: c.ResponseChan,
			}

		}
	}

}

func (c *client) handleError(errMessage string) error {
	response := fmt.Sprintf("-ERR %s\r\n", errMessage)
	_, err := c.Writer.Write([]byte(response))
	if err != nil {
		return err
	}
	err = c.Writer.Flush()
	if err != nil {
		return err
	}
	return nil
}
