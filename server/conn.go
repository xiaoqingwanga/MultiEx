package server

import (
	"MultiEx/util"
	"net"
)

type wrappedconn struct {
	id   string
	log  util.PrefixLogger
	conn net.Conn
}

type listener struct {
	conns chan *wrappedconn
}

func listen(port string) (l *listener, e error) {
	// old style listener
	oL, err := net.Listen("tcp", port)
	if err != nil {
		panic(err)
	}
	l = &listener{
		conns: make(chan *wrappedconn),
	}
	go func() {
		for {
			c, err := oL.Accept()
			if err != nil {
				// Todo: judge this error is handleable?
				continue
			}
			// wrap connection
			wC := &wrappedconn{
				conn: c,
				log:  util.NewPrefixLogger("conn"),
			}
		}
	}()
}
