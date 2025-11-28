package main

import (
    "fmt"
    "log"
    "net"

    "github.com/aabbcc333/httpfromtcp/internal/request"
)

func main() {
    listener, err := net.Listen("tcp", ":42069")
    if err != nil {
        log.Fatal("error", err)
    }

    log.Println("listening on :42069")

    for {
        log.Println("listening on :42069 entering loop")
        conn, err := listener.Accept()
        if err != nil {
            log.Fatal("accept error:", err)
        }

        log.Println("accepted connection from", conn.RemoteAddr())

        r, err := request.RequestFromReader(conn)
        if err != nil {
            log.Fatal("request parse error:", err)
        }

        fmt.Printf("Request line:\n")
        fmt.Printf("- Method: %s\n", r.RequestLine.Method)
        fmt.Printf("- Target: %s\n", r.RequestLine.RequestTarget)
        fmt.Printf("- Version: %s\n", r.RequestLine.HttpVersion)

        fmt.Printf("Headers:\n")
        r.Headers.ForEach(func(n, v string) {
            fmt.Printf("- %s: %s\n", n, v)
        })
    }
}