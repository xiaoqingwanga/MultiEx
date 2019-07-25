package server

import (
	"MultiEx/msg"
	"io"
	"net"
	"time"
)

// Client represent a MultiEx client in server.
type Client struct {
	ID        string
	Conn      Conn
	Ports     []string
	Listeners []net.Listener
	Proxies   chan Conn
	LastPing  *time.Time
}

func (client *Client) Close() {
	defer func() {
		if r := recover(); r != nil {
			client.Conn.Warn("unexpected error: %v", r)
		}
	}()
	client.Conn.Info("client closing...")
	client.LastPing = nil
	client.Conn.Close()
	// close channel,dont accept new proxy
	close(client.Proxies)
	for c := range client.Proxies {
		msg.WriteMsg(client.Conn, msg.CloseProxy{})
		c.Close()
	}
	for _, l := range client.Listeners {
		l.Close()
	}
	client.Conn.Info("close finished")
}

func (client *Client) AcceptCmd(reg *ClientRegistry) {
	go func() {
		ticker := time.Tick(time.Second * 31)
		for {
			select {
			case <-ticker:
				if client.LastPing == nil {
					client.Conn.Error("client already closed, stop ticker for ping")
					return
				}
				if time.Now().Sub(*client.LastPing) > time.Minute {
					client.Conn.Warn("client not ping too long")
					client.Close()
					reg.Unregister(client.ID)
				}
			}
		}
	}()
	for {
		m, e := msg.ReadMsg(client.Conn)
		if e != nil {
			client.Conn.Warn("%s when read message", e)
			client.Close()
			// Maybe denial of service attack
			break
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
			l, e := net.Listen("tcp", ":"+port)
			if e != nil {
				client.Conn.Warn("port %s is in use", port)
				msg.WriteMsg(client.Conn, msg.PortInUse{Port: port})
				return
			}
			client.Listeners = append(client.Listeners, l)
			for {
				c, e := l.Accept()
				if e != nil {
					client.Conn.Warn("listener at %s closed", port)
					break
				}
				client.Conn.Info("remote host:%s is coming", c.RemoteAddr().String())
				go handlePublic(port, c, client)
			}

		}(p)
	}
}

func handlePublic(port string, c net.Conn, client *Client) {
	defer func() {
		if r := recover(); r != nil {
			client.Conn.Error("fatal when handle public conn:%v", r)
		}
	}()

	var proxy Conn
	var i int
	for success := false; i < 15 && !success; i++ {

		client.Conn.Info("try to get proxy connection,times:%d", i+1)
		select {
		case proxy = <-client.Proxies:
			// A new proxy
			msg.WriteMsg(client.Conn, msg.NewProxy{})
			e := msg.WriteMsg(proxy, msg.ForwardInfo{Port: port})
			if e == nil {
				success = true
				break
			}

		default:
			client.Conn.Info("there isn't any proxy available, ask client connect")
			msg.WriteMsg(client.Conn, msg.NewProxy{})
			select {
			case proxy = <-client.Proxies:
				msg.WriteMsg(client.Conn, msg.NewProxy{})
				e := msg.WriteMsg(proxy, msg.ForwardInfo{Port: port})
				if e == nil {
					success = true
					break
				}
			case <-time.After(time.Second * 20):
				client.Conn.Error("wait for 20 seconds, and there isn't any proxy available still")
				client.Conn.Error("cannot get proxy, client to be closed")
				client.Close()
				return
			}
		}
	}

	if i == 15 {
		client.Conn.Error("cannot get proxy, client to be closed")
		client.Close()
		return
	}

	proxy.AddPrefix("remote-" + c.RemoteAddr().String())
	proxy.Info("proxy selected, forward start")

	defer func() {
		client.Conn.Info("forward finished, public visitor:%s", c.RemoteAddr().String())
		proxy.Close()
		c.Close()
	}()
	// begin transfer data between them.
	go io.Copy(proxy, c)
	io.Copy(c, proxy)
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

// Register register client
func (registry *ClientRegistry) Unregister(id string) (oClient *Client) {
	oClient, ok := (*registry)[id]
	if ok {
		delete(*registry, id)
		return
	}
	return
}
