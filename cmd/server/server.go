package main

import (
	"bufio"
	"fmt"
	"net"
	"net/textproto"
	"sync"
)

type Endpoint struct {
	l net.Listener
	c net.Conn
	w sync.WaitGroup
	d chan string
}


func NewEndpoint(addr string) *Endpoint {
	e := &Endpoint{}
	l, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	e.l = l
	e.d = make(chan string)
	e.w.Add(1)
	go func() {
		defer e.w.Done()
		for {
			c, err := l.Accept()
			if err != nil {
				fmt.Println("closed listener")
				break
			}
			fmt.Println("accepted", c)
			e.c = c
			r := textproto.NewReader(bufio.NewReader(c))
			for {
				l, err := r.ReadLine()
				if err != nil {
					fmt.Println("closed socket")
					c.Close()
					e.c = nil
					break
				}
				e.d <- l
			}
		}
	}()
	return e
}

func (e *Endpoint) Stop() {
	fmt.Println("stop endpoint")
	e.l.Close()
	if e.c != nil {
		e.c.Close()
	}
	e.w.Wait()
}

