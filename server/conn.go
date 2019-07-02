package server

import (
	"MultiEx/util"
	"math/rand"
	"net"
	"time"
)

// Conn represents a connection with logger.
type Conn interface {
	AddPrefix(pfx string)
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
	Write(b []byte) (n int, err error)
	Read(b []byte) (n int, err error)
	Close() error
	GetID() string
}

type wrappedconn struct {
	ID string
	util.PrefixLogger
	net.Conn
}

func (wc wrappedconn) GetID() string  {
	return wc.ID
}

type listener struct {
	conns chan Conn
}

func listen(port string) (l *listener) {
	// old style listener
	oL, err := net.Listen("tcp", port)
	if err != nil {
		panic(err)
	}
	l = &listener{
		conns: make(chan Conn),
	}
	go func() {
		for {
			c, err := oL.Accept()
			if err != nil {
				util.Error("listener %v is closed",oL)
				break
			}
			// wrap connection
			wC := &wrappedconn{
				ID:           string(time.Now().Unix()) + string(rand.Int31n(10)),
				Conn:         c,
				PrefixLogger: util.NewPrefixLogger("conn"),
			}
			wC.AddPrefix(wC.ID)
			l.conns <- wC
		}
	}()
	return
}
