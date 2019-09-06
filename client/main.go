package client

import (
	"MultiEx/log"
	"MultiEx/msg"
	"flag"
	"io"
	"net"
	"strings"
	"time"
)

type options struct {
	logTo      string
	logLevel   string
	remotePort string
	token      string
	portMap    map[string]string
}

var PortMap map[string]string
var ClientID string

var pingFlag bool

/**
控制连接终端后，重试三次
*/
const RETRY_LIMIT = 3;

var retryCount = 0;

func Main() {
	options := option()

	log.Init(options.logLevel, options.logTo)

	PortMap = make(map[string]string)
	PortMap = options.portMap

	//work(options.remotePort, options.token)
	work(options.remotePort, options.token)
}

func option() options {
	logTo := flag.String("logTo", "stdout", "the location where logs save. Empty value and stdout have special meaning")
	logLevel := flag.String("logLevel", "INFO", "the log level of this program.")
	remotePort := flag.String("remotePort", "", "the public server ip:port listening for MultiEx client.")
	token := flag.String("token", "", "Token is the credential client should hold to connect server.Server doesn't have token default.")
	ports := flag.String("portMap", "2222-22", "Port map represent mapping between host. "+
		"e.g. '2222-22' represents expose local port 22 at public port 2222. Multi mapping split by comma.")
	flag.Parse()

	portMap := make(map[string]string)
	pairs := strings.Split(*ports, ",")
	for _, p := range pairs {
		mapping := strings.Split(p, "-")
		portMap[mapping[0]] = mapping[1]
	}
	return options{
		logTo:      *logTo,
		logLevel:   *logLevel,
		remotePort: *remotePort,
		token:      *token,
		portMap:    portMap,
	}
}

func work(remote string, token string) {

	ctrl, ports := connect(remote, token)
	if ctrl == nil {
		return
	}
	msg.WriteMsg(ctrl, msg.NewClient{Token: token, Forwards: ports})
	for {
		m, e := msg.ReadMsg(ctrl)
		if e != nil {
			log.Error("control connection die, %v.", e)
			log.Error("maybe wrong token")
			retryCount++;
			ctrl = nil
			for retryCount < RETRY_LIMIT && ctrl == nil {
				if retryCount != 0 {
					log.Info("connect fail.try again 2 seconds later")
					time.Sleep(time.Second * 2)
				}
				log.Info("try to reconnect %d", retryCount)
				ctrl, ports = connect(remote, token)
			}
			if ctrl != nil {
				msg.WriteMsg(ctrl, msg.NewClient{Token: token, Forwards: ports})
				continue
			} else {
				log.Info("cannot connect server,exit")
				return
			}
		}
		retryCount = 0
		switch nm := m.(type) {
		case *msg.ReNewClient:
			ClientID = nm.ID
			go ping(ctrl)
		case *msg.Pong:
			log.Info("server pong")
			pingFlag = false
		case *msg.PortInUse:
			log.Warn("port %s is in use. %s -> %s not take effect", nm.Port, nm.Port, PortMap[nm.Port])
			delete(PortMap, nm.Port)
			if len(PortMap) == 0 {
				log.Warn("no port mapping available,exit")
				return
			}
		case *msg.NewProxy:
			log.Info("receive NewProxy cmd")
			p, e := net.Dial("tcp", remote)
			if e != nil {
				log.Error("cannot dial remote,%v", e)
				break
			}
			msg.WriteMsg(p, msg.NewProxy{ClientID: ClientID})
			go forward(p)
		}
	}
}

func connect(remote, token string) (conn net.Conn, ports []string) {
	log.Info("attempt to connect '%s' with token '%s' ...", remote, token)
	conn, e := net.DialTimeout("tcp", remote, time.Second*5)
	if e != nil {
		log.Error("%v", e)
		conn = nil
		return
	}
	log.Info("connect server success")

	for p := range PortMap {
		ports = append(ports, p)
	}
	return
}

func ping(c net.Conn) {
	for {
		ticker := time.Tick(time.Second * 10)
		select {
		case <-ticker:
			if pingFlag {
				log.Warn("your network is busy")
			}
			log.Info("ping server")
			e := msg.WriteMsg(c, msg.Ping{})
			if e != nil {
				break
			}
			pingFlag = true
		}
	}
}

func forward(c net.Conn) {
	defer func() {
		c.Close()
	}()
	m, e := msg.ReadMsg(c)
	if e != nil {
		log.Warn("proxy connection die")
		return
	}

	nm, ok := m.(*msg.ForwardInfo)
	if !ok {
		log.Warn("remote server seems insane...")
		return
	}

	log.Info("forwarding...")

	lc, e := net.Dial("tcp", ":"+PortMap[nm.Port])
	if e != nil {
		log.Warn("dial local port fail, %v", e)
		c.Close()
		return
	}
	defer lc.Close()
	go io.Copy(lc, c)
	io.Copy(c, lc)
	log.Info("forward finished")
	return
}
