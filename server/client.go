package server

import (
	"MultiEx/msg"
	"io"
	"net"
	"time"
)

// Client represent a MultiEx client in server.
type Client struct {
	ID       string
	Conn     Conn
	Ports    []string
	Proxies  chan Conn
	LastPing *time.Time
}

func (client *Client) stop() {
	client.Conn.Warn("close control connection")
	client.LastPing = nil
	client.Conn.Close()
	close(client.Proxies)
	for c := range client.Proxies {
		msg.WriteMsg(client.Conn, msg.CloseProxy{})
		c.Close()
	}
}

func (client *Client) AcceptCmd() {
	go func() {
		ticker := time.Tick(time.Second * 30)
		for {
			select {
			case <-ticker:
				if client.LastPing == nil {
					client.Conn.Error("client closed, stop ticker for ping")
					break
				}
				if time.Now().Sub(*client.LastPing) > time.Minute {
					client.Conn.Warn("client not ping too long")
					client.stop()
				}
			}
		}
	}()
	for {
		m, e := msg.ReadMsg(client.Conn)
		if e != nil {
			client.Conn.Warn("error when read message: %v", e)
			// Maybe denial of service attack
			switch e {
			case io.EOF, io.ErrUnexpectedEOF:
				break
			default:
				continue
			}
		}
		switch m.(type) {
		case *msg.Ping:
			client.Conn.Info("ping!")
			now := time.Now()
			client.LastPing = &now
			msg.WriteMsg(client.Conn, msg.Pong{})
		}
	}
}

func (client *Client) StartListener() {
	for _, p := range client.Ports {
		go func(port string) {
			l, e := net.Listen("tcp", p)
			if e != nil {
				client.Conn.Warn("%s is in use", p)
				msg.WriteMsg(client.Conn, msg.PortInUse{Port: p})
				return
			}
			for {
				c, e := l.Accept()
				if e != nil {
					client.Conn.Warn("listener at %s closed", p)
					break
				}
				client.Conn.Info("remote host:%s is coming", c.RemoteAddr().String())
				go handlePublic(port, c, client)
			}

		}(p)
	}
}

func handlePublic(port string, c net.Conn, client *Client) {
	var proxy Conn
	var i int
	for ; i < 15; i++ {
		client.Conn.Info("try to get proxy connection,times:%d", i+1)
		select {
		case proxy = <-client.Proxies:
			// Must write twice to test if connection close
			msg.WriteMsg(proxy, msg.ActivateProxy{})
			e := msg.WriteMsg(proxy, msg.ForwardInfo{Port: port})
			if e == nil {
				break
			}
			msg.WriteMsg(client.Conn, msg.NewProxy{})
		default:
			client.Conn.Info("there isn't any proxy available, ask client connect")
			msg.WriteMsg(client.Conn, msg.NewProxy{})
			select {
			case proxy = <-client.Proxies:
				msg.WriteMsg(proxy, msg.ActivateProxy{})
				e := msg.WriteMsg(proxy, msg.ForwardInfo{Port: port})
				if e == nil {
					break
				}
				msg.WriteMsg(client.Conn, msg.NewProxy{})
			case <-time.After(time.Second * 10):
				client.Conn.Error("wait for 10 seconds, and there isn't any proxy available still")
				client.Conn.Error("cannot get proxy, client to be closed")
				client.stop()
				return
			}
		}
	}

	if i == 15 {
		client.Conn.Error("cannot get proxy, client to be closed")
		client.stop()
		return
	}

	proxy.AddPrefix("client-" + client.ID)
	proxy.AddPrefix("remote-" + c.RemoteAddr().String())
	proxy.Info("proxy selected, begin transfer data")

	defer func() {
		client.Conn.Info("remote host:%s data transfer finished.", c.RemoteAddr().String())
		proxy.Close()
		c.Close()
	}()
	// begin transfer data between them.
	go io.Copy(c, proxy)
	io.Copy(proxy, c)
	return
}

// ClientRegistry is a place storing clients.
type ClientRegistry map[string]*Client

// Register register client
func (registry *ClientRegistry) Register(id string, client *Client) (oClient *Client) {
	oClient, ok := (*registry)[id]
	if ok {
		return
	}
	(*registry)[id] = client
	return
}
