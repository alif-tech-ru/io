package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"strings"
)

func main() {
	host := "0.0.0.0"
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "9999"
	}
	err := start(fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		log.Fatal(err)
	}
}

func start(addr string) (err error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("can't listen %s: %w", addr, err)
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		log.Print("accept connection")
		if err != nil {
			log.Printf("can't accept: %v", err)
			continue
		}
		log.Print("handle connection")
		handleConn(conn)
	}
}

// apache
// nginx
// IIS
// Apache Tomcat, Jetty
// go http server

// Request-Line\r\n
// Headers\r\n
// Headers\r\n
// \r\n
// Body
func handleConn(conn net.Conn) {
	defer conn.Close()

	log.Print("read request to buffer")
	const maxHeaderSize = 4096
	reader := bufio.NewReaderSize(conn, maxHeaderSize)
	writer := bufio.NewWriter(conn)
	counter := 0
	buf := [maxHeaderSize]byte{}
	// naive header limit
	for {
		if counter == maxHeaderSize {
			log.Printf("too long request header")
			writer.WriteString("HTTP/1.1 413 Payload Too Large\r\n")
			writer.WriteString("Content-Length: 0\r\n")
			writer.WriteString("Connection: close\r\n")
			writer.WriteString("\r\n")
			writer.Flush()
			return
		}

		read, err := reader.ReadByte()
		if err != nil {
			log.Printf("can't read request line: %v", err)
			writer.WriteString("HTTP/1.1 400 Bad Request\r\n")
			writer.WriteString("Content-Length: 0\r\n")
			writer.WriteString("Connection: close\r\n")
			writer.WriteString("\r\n")
			writer.Flush()
			return
		}
		buf[counter] = read
		counter++

		if counter < 4 {
			continue
		}

		if string(buf[counter-4:counter]) == "\r\n\r\n" {
			break
		}
	}

	log.Print("headers found")
	headersStr := string(buf[:counter - 4])

	headers := make(map[string]string) // TODO: в оригинале map[string][]string
	requestHeaderParts := strings.Split(headersStr, "\r\n")

	log.Print("parse request line")
	requestLine := requestHeaderParts[0]
	log.Printf("request line: %s", requestLine)

	log.Print("parse headers")
	for _, headerLine := range requestHeaderParts[1:] {
		headerParts := strings.SplitN(headerLine, ": ", 2)
		headers[strings.TrimSpace(headerParts[0])] = strings.TrimSpace(headerParts[1]) // TODO: are we allow empty header?
	}
	log.Printf("headers: %v", headers)

	html := fmt.Sprintf(`<!doctype html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport"
          content="width=device-width, user-scalable=no, initial-scale=1.0, maximum-scale=1.0, minimum-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="ie=edge">
    <title>Document</title>
</head>
<body>
    <h1>Hello from golang %s</h1>
</body>
</html>`, runtime.Version())

	log.Print("send response")
	writer.WriteString("HTTP/1.1 200 OK\r\n")
	writer.WriteString(fmt.Sprintf("Content-Length: %d\r\n", len(html)))
	writer.WriteString("Connection: close\r\n")
	writer.WriteString("\r\n")
	writer.WriteString(html)
	writer.Flush()

	log.Print("done")
	return
}
