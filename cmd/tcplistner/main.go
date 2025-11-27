package main

import (
	//"bytes"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	//"os"
)

func getLinesChannel(f io.ReadCloser) <-chan string{
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
}

func main() {
  listner, err := net.Listen("tcp",":42069")
  if err != nil {
	log.Fatal("error","error",err)
  }

  for{
	conn,err := listner.Accept()
	if err != nil{
		
	log.Fatal("error","error",err)
	}
	for line := range getLinesChannel(conn){
		fmt.Printf("read:%s\n",line)
	}
  }
}
