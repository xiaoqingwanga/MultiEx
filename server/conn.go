package server

import (
	"MultiEx/log"
	"math/rand"
	"net"
	"os"
	"strconv"
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
	RemoteAddr() net.Addr
	ReplacePrefix(old string, new string)
}

type wrappedconn struct {
	ID string
	log.PrefixLogger
	net.Conn
}

func (wc wrappedconn) GetID() string {
	return wc.ID
}

type listener struct {
	conns chan Conn
}

func listen(port string, reg *ClientRegistry) (l *listener) {
	// old style listener
	oL, err := net.Listen("tcp", port)
	if err != nil {
		panic(err)
	}
	l = &listener{
		conns: make(chan Conn),
	}
	log.Info("listen at %s", port)
	go func() {
		for {
			c, err := oL.Accept()
			if err != nil {
				log.Error("MultiEx client listener closed,%v", err)
				stopApp(reg)
				return
			}
			// wrap connection
			wC := &wrappedconn{
				ID:           strconv.Itoa(int(time.Now().Unix())) + strconv.Itoa(rand.Intn(10)),
				Conn:         c,
				PrefixLogger: log.NewPrefixLogger(),
			}
			wC.AddPrefix("conn-" + wC.ID)
			l.conns <- wC
		}
	}()
	return
}

func stopApp(reg *ClientRegistry) {
	log.Error("exit app")
	for _, c := range *reg {
		c.Close()
	}
	os.Exit(1)
}
