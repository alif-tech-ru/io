package main

import (
	"bufio"
	"io/ioutil"
	"net"
	"strings"
	"testing"
	"time"
)

func Test_server(t *testing.T) {
	// naive approach
	go func() {
		err := start("localhost:9999")
		if err != nil {
			t.Fatalf("can't start server: %v", err)
		}
	}()
	// FIXME: hardcoded approach
	time.Sleep(1_000_000_000)
	conn, err := net.Dial("tcp", "localhost:9999")
	if err != nil {
		t.Fatalf("can't connect to server: %v", err)
	}
	defer conn.Close()
	writer := bufio.NewWriter(conn)
	writer.WriteString("GET / HTTP/1.1\r\n")
	writer.WriteString("Host: localhost\r\n")
	writer.WriteString("\r\n")
	writer.Flush()
	bytes, err := ioutil.ReadAll(conn)
	if err != nil {
		t.Fatalf("can't read response from server: %v", err)
	}
	response := string(bytes)
	if !strings.Contains(response, "200 OK") {
		t.Fatalf("non-success response: %s", response)
	}
}
