package server

import (
	"MultiEx/msg"
	"MultiEx/registry"
	"io"
	"math/rand"
	"time"
)

// Client represent a MultiEx client in server.
type Client struct {
	ID       string
	Conn     Conn
	Proxies  []Conn
	LastPing *time.Time
}

func (client *Client) acceptCmd() {
	go func() {
		ticker := time.Tick(time.Second * 30)
		for {
			select {
			case <-ticker:
				if client.LastPing==nil{
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
			case io.EOF,io.ErrUnexpectedEOF:
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
			msg.WriteMsg(client.Conn,msg.Pong{})
		}
	}
}

func (client *Client) startListener() {

}

func (client *Client) stop() {
	client.Conn.Warn("client not ping too long, close")
	client.LastPing = nil
	msg.WriteMsg(client.Conn, msg.GResponse{Msg: "control connection close for some reason"})
	client.Conn.Close()
	for _, c := range client.Proxies {
		msg.WriteMsg(client.Conn, msg.GResponse{Msg: "proxy connection close for some reason"})
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
				client := &Client{
					ID:       string(time.Now().Unix() + rand.Int63n(10)),
					Conn:     c,
					Proxies:  make([]Conn, 0),
					LastPing: time.Now(),
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
