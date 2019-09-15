package client

import (
	"MultiEx/log"
	"MultiEx/msg"
	"MultiEx/util"
	"flag"
	"fmt"
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

var ctrlLog log.PrefixLogger
var proxyLog log.PrefixLogger

var PortMap map[string]string
var ClientID string
var counterMap map[string]*util.Count
var retryCount int
var inUsePortCount int

func Main() {
	options := option()
	log.Init(options.logLevel, options.logTo)
	proxyLog.AddPrefix("proxy")

	counterMap = make(map[string]*util.Count)

	PortMap = make(map[string]string)
	PortMap = options.portMap

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

	//logTo := "stdout"
	//logLevel := "INFO"
	//remotePort := "182.61.18.71:8070"
	//token := "a"
	//ports := "8444-8444"

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

	ctrl, ports := dial(remote, token)
	if ctrl == nil {
		return
	}
	msg.WriteMsg(ctrl, msg.NewClient{Token: token, Forwards: ports})
	for {
		m, e, r := msg.ReadMsg(ctrl)
		if r {
			continue
		}
		if e != nil || inUsePortCount == len(PortMap) {
			if e != nil {
				ctrlLog.Error("error when read cmd,%v", e)
			}
			ctrl = nil
			inUsePortCount = 0
			template := "after %d seconds try to reconnect %s time"
			for retryCount < 3 && ctrl == nil {
				var tip string
				var t time.Duration
				switch retryCount {
				case 0:
					tip = fmt.Sprintf(template, 5, "first")
					t = 5 * time.Second
				case 1:
					tip = fmt.Sprintf(template, 60, "second")
					t = 60 * time.Second
				case 2:
					tip = fmt.Sprintf(template, 120, "third")
					t = 120 * time.Second
				}
				log.Info(tip)
				time.Sleep(t)
				ctrl, _ = dial(remote, token)
				if ctrl != nil &&
					msg.WriteMsg(ctrl, msg.NewClient{Token: token, Forwards: ports}) != nil {
					ctrl = nil
				}
				retryCount++
			}
		}
		if ctrl == nil {
			log.Info("cannot connect server,exit")
			return
		}
		switch nm := m.(type) {
		case *msg.ReNewClient:
			ctrlLog = log.NewPrefixLogger("ctrl-" + nm.ID)
			ctrlLog.Info("successfully connect server")
			ClientID = nm.ID
			var count util.Count
			counterMap[ClientID] = &count
			go ping(ctrl, ClientID)
		case *msg.Pong:
			c, ok := counterMap[ClientID]
			if ok {
				c.Dec()
			}
		case *msg.PortInUse:
			inUsePortCount++
			log.Warn("server port %s is in use. mapping %s -> %s not take effect", nm.Port, nm.Port, PortMap[nm.Port])
		case *msg.NewProxy:
			//not elegant. when client receive newproxy represents client and server communicate perfectly
			retryCount = 0

			ctrlLog.Info("receive NewProxy cmd")
			p, e := net.Dial("tcp", remote)
			if e != nil {
				ctrlLog.Error("error when new a proxy,%v", e)
				break
			}
			msg.WriteMsg(p, msg.NewProxy{ClientID: ClientID})
			go forward(p)
		}
	}
}

func dial(remote, token string) (conn net.Conn, ports []string) {
	log.Info("attempt to connect '%s' with token '%s'", remote, token)
	conn, e := net.DialTimeout("tcp", remote, time.Second*15)
	if e != nil {
		log.Error("dial server fail, %v", e)
		conn = nil
		return
	}
	log.Info("dial server success")

	for p := range PortMap {
		ports = append(ports, p)
	}
	return
}

func ping(conn net.Conn, clientId string) {
	for {
		counter, ok := counterMap[clientId]
		//log.Info("counter:%d", counter.Get())
		if !ok || counter.Get() > 4 {
			log.Info("no heart beat for a long time with client id:%s, stop ping", clientId)
			return
		}
		ticker := time.Tick(time.Second * 10)
		select {
		case <-ticker:
			e := msg.WriteMsg(conn, msg.Ping{})
			if e != nil {
				counter.IncN(3)
			}
			counter.Inc()
		}
	}
}

func forward(c net.Conn) {
	defer func() {
		c.Close()
	}()
	m, e, _ := msg.ReadMsg(c)
	if e != nil {
		proxyLog.Warn("error when read cmd,%v", e)
		return
	}

	nm, ok := m.(*msg.ForwardInfo)
	if !ok {
		proxyLog.Warn("remote server seems insane...")
		return
	}

	proxyLog.Info("start a forward")

	lc, e := net.Dial("tcp", ":"+PortMap[nm.Port])
	if e != nil {
		proxyLog.Warn("dial local port fail, %v", e)
		c.Close()
		return
	}
	defer lc.Close()
	go io.Copy(lc, c)
	io.Copy(c, lc)
	proxyLog.Info("a forward finished")
	return
}
