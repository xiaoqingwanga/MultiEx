package client

import (
	"MultiEx/msg"
	"MultiEx/util"
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

func Main() {
	options := option()

	util.Initlog(options.logLevel, options.logTo)

	PortMap = make(map[string]string)
	PortMap = options.portMap

	//work(options.remotePort, options.token)
	work(":8070", options.token)
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
	ctrl, e := net.Dial("tcp", remote)
	if e != nil {
		util.Info("%v", e)
		return
	}
	util.Info("connect server success")

	var ports []string
	for p, _ := range PortMap {
		ports = append(ports, p)
	}
	msg.WriteMsg(ctrl, msg.NewClient{Token: token, Forwards: ports})
	for {
		m, e := msg.ReadMsg(ctrl)
		if e != nil {
			util.Error("server die, %v.maybe wrong token", e)
			return
		}
		switch nm := m.(type) {
		case *msg.ReNewClient:
			ClientID = nm.ID
			go ping(ctrl)
		case *msg.Pong:
			util.Info("server pong")
			pingFlag = false
		case *msg.PortInUse:
			util.Warn("port %s is in use. %s -> %s not take effect", nm.Port, nm.Port, PortMap[nm.Port])
			delete(PortMap, nm.Port)
			if len(PortMap)==0{
				util.Warn("no port mapping available,exit")
				return
			}
		case *msg.NewProxy:
			p, e := net.Dial("tcp", remote)
			if e != nil {
				util.Error("cannot dial remote:%v", e)
				break
			}
			go proxyWork(p)
		}
	}
}

func ping(c net.Conn) {
	for {
		ticker := time.Tick(time.Second * 5)
		select {
		case <-ticker:
			if pingFlag {
				util.Warn("seems server die...")
			}
			util.Info("ping server")
			e := msg.WriteMsg(c, msg.Ping{})
			if e != nil {
				break
			}
			pingFlag = true
		}
	}
}

func proxyWork(c net.Conn) {
	defer func() {
		util.Info("a proxy stop")
		c.Close()
	}()
	msg.WriteMsg(c, msg.NewProxy{ClientID: ClientID})
	m, _ := msg.ReadMsg(c)
	m, e := msg.ReadMsg(c)
	util.Info("remote server ask me start a proxy")
	if e != nil {
		util.Warn("proxy connection die")
		return
	}
	nm, ok := m.(*msg.ForwardInfo)
	if !ok{
		util.Warn("remote server seems insane...")
		return
	}

	lc,e := net.Dial("tcp",":"+PortMap[nm.Port])
	if e!=nil{
		util.Warn("dial local port fail, %v",e)
		c.Close()
		return
	}
	defer lc.Close()
	go io.Copy(c,lc)
	io.Copy(lc,c)
	return
}
