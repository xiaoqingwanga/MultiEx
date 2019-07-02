package server

import (
	"MultiEx/msg"
	"MultiEx/registry"
	"io"
	"math/rand"
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

func (client *Client) acceptCmd() {
	go func() {
		ticker := time.Tick(time.Second * 30)
		for {
			select {
			case <-ticker:
				if client.LastPing == nil {
					client.Conn.Error("client closed, stop ticker for ping")
				}
				if time.Now().Sub(*client.LastPing) > time.Minute {
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

func (client *Client) startListener() {
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

				var proxy Conn
				var i int
				for ; i < 10*2; i++ {
					client.Conn.Info("try to get proxy connection,times:%d",i+1)
					select {
					case proxy = <-client.Proxies:
						msg.WriteMsg(proxy, msg.ActivateProxy{})
						e = msg.WriteMsg(proxy, msg.ActivateProxy{})
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
							e = msg.WriteMsg(proxy, msg.ActivateProxy{})
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

				if i==10*2{
					client.Conn.Error("cannot get proxy, client to be closed")
					client.stop()
					return
				}

				proxy.AddPrefix("client-"+client.ID)
				proxy.Info("proxy selected, begin transfer data")


				// begin transfer data between them.

			}

		}(p)
	}
}

func (client *Client) stop() {
	client.Conn.Warn("client not ping too long, close")
	client.LastPing = nil
	msg.WriteMsg(client.Conn, msg.GResponse{Msg: "control connection close for some reason"})
	client.Conn.Close()
	close(client.Proxies)
	for c := range client.Proxies {
		msg.WriteMsg(client.Conn, msg.CloseProxy{})
		c.Close()
	}
}

// HandleClient accept client control connection,proxy connection
func HandleClient(token string, port string, registry registry.ClientRegistry) {
	// Listen
	l := listen(port)

	// Get and Handle new connection
	for c := range l.conns {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					c.Warn("%v", r)
				}
			}()

			c.Info("start read message")
			m, e := msg.ReadMsg(c)
			if e != nil {
				c.Warn("cannot read message from connection")
				msg.WriteMsg(c, msg.GResponse{Msg: "servcer cannot identify your connection"})
				c.Close()
				return
			}
			c.Info("successfully read message")
			switch nM := m.(type) {
			case *msg.NewClient:
				if token != (*nM).Token {
					c.Warn("wrong token taken from client.Close")
					msg.WriteMsg(c, msg.GResponse{Msg: "cannot identify your token"})
					c.Close()
					return
				}

				now := time.Now()
				client := &Client{
					ID:       string(time.Now().Unix() + rand.Int63n(10)),
					Conn:     c,
					Ports:    nM.Forwards,
					Proxies:  make(chan Conn, 10),
					LastPing: &now,
				}
				c.AddPrefix("client-" + client.ID)
				registry.Register(client.ID, client)
				c.Info("client registered")
				msg.WriteMsg(c, msg.ReNewClient{
					ID: client.ID,
				})
				go client.acceptCmd()
				go client.startListener()
			}
		}()
	}
}
