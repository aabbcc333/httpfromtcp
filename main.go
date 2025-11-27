package main

import (
	//"bytes"
	//"bytes"
	"fmt"
	//"io"
	"log"
	"net"

	request "github.com/aabbcc333/httpfromtcp/internal"
	//"os"
)

/*func getLinesChannel(f io.ReadCloser) <-chan string{
	out := make(chan string,1)
	buffer := make([]byte, 8)
	var str []byte
	go func(){
		defer f.Close()
		defer close(out)
		for{
			n,err := f.Read(buffer)


			chunk := buffer[:n]
			for len(chunk)>0{
             index := bytes.IndexByte(chunk,'\n')
			 if index == -1{
				str = append(str, chunk...)
				chunk = chunk[:0]
				continue
			 }
			 strPart := chunk[:index]
			 str = append(str,strPart...)
			 out <- string(str)
			 str = str[:0]
			 chunk = chunk[index+1:]
			}

			if err == io.EOF{
				if len(str)>0{
					out <- string(str)
				}
				return
			}
			if err != nil {
				fmt.Printf("error %s", err)
				return
			}
		}


	}()
	return out
}*/


func main() {
  listner, err := net.Listen("tcp",":42069")
  if err != nil {
	log.Fatal("error","error",err)
  }

  log.Println("listening on :42069")

  for{
	log.Println("listening on :42069 etnering loop")
	conn,err := listner.Accept()
	if err != nil{
		
	log.Fatal("error","error",err)
	}
	
	
	r,err := request.RequestFromReader(conn)
	if err != nil{
		log.Fatal("error", "error", err)
	}
	fmt.Printf("Request line: \n")
	fmt.Printf("-Method: %s\n", r.RequestLine.Method)
	fmt.Printf("-Target: %s\n", r.RequestLine.RequestTarget)
	fmt.Printf("-Version: %s\n", r.RequestLine.HttpVersion)
  }
}
