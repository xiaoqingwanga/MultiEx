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

var PortMap map[string]string
var ClientID string

var pingFlag util.Count

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

	ctrl, ports := dial(remote, token)
	if ctrl == nil {
		return
	}
	log.Info("attempt to connect...")
	msg.WriteMsg(ctrl, msg.NewClient{Token: token, Forwards: ports})
	for {
		m, e := msg.ReadMsg(ctrl)
		if e != nil {
			log.Error("%v", e)
			var retryCount int
			ctrl = nil
			template := "try to reconnect %s times, after %d seconds"
			for retryCount < 3 && ctrl == nil {
				var tip string
				var t time.Duration
				switch retryCount {
				case 0:
					tip = fmt.Sprintf(template, "1st", 3)
					t = 5 * time.Second
				case 1:
					tip = fmt.Sprintf(template, "2nd", 60)
					t = 30 * time.Second
				case 2:
					tip = fmt.Sprintf(template, "3rd", 120)
					t = 60 * time.Second
				}
				log.Info(tip)
				time.Sleep(t)
				ctrl, _ = dial(remote, token)
				retryCount++
			}
			if ctrl != nil {
				msg.WriteMsg(ctrl, msg.NewClient{Token: token, Forwards: ports})
				continue
			} else {
				log.Info("cannot connect server,exit")
				return
			}
		}
		switch nm := m.(type) {
		case *msg.ReNewClient:
			log.Info("connect server success, and client get a id:" + nm.ID)
			ClientID = nm.ID
			go ping(ctrl, nm.ID)
		case *msg.Pong:
			pingFlag.Dec()
		case *msg.PortInUse:
			log.Warn("server port %s is in use. mapping %s -> %s not take effect", nm.Port, nm.Port, PortMap[nm.Port])
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

func dial(remote, token string) (conn net.Conn, ports []string) {
	log.Info("attempt to connect '%s' with token '%s' ...", remote, token)
	conn, e := net.DialTimeout("tcp", remote, time.Second*5)
	if e != nil {
		log.Error("%v", e)
		conn = nil
		return
	}
	log.Info("dial server success")

	for p := range PortMap {
		ports = append(ports, p)
	}
	return
}

func ping(c net.Conn, clientId string) {
	var fail bool
	for {
		if pingFlag.Get() > 2 || fail {
			log.Info("server no heart beat for a long time" + ", and current client id:" + clientId)
			return
		}
		ticker := time.Tick(time.Second * 10)
		select {
		case <-ticker:
			e := msg.WriteMsg(c, msg.Ping{})
			if e != nil {
				fail = true
			}
			pingFlag.Inc()
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
