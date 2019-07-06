package main

import (
	"fmt"
	"net"
)

func main()  {
	l,_ := net.Listen("tcp",":1800")
	for  {
		c,_ := l.Accept()
		go func() {
			for  {
				bytes := make([]byte,10)
				c.Read(bytes)
				fmt.Println(string(bytes))
			}
		}()
		go func() {
			for  {
				bytes := []byte("a")
				c.Write(bytes)
			}
		}()
	}
}
