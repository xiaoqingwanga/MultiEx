package server

import (
	"MultiEx/msg"
	"MultiEx/util"
	"math/rand"
	"strconv"
	"time"
)

// HandleClient accept client control connection,proxy connection
func HandleClient(port string, token string, reg ClientRegistry) {
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

			m, e := msg.ReadMsg(c)
			if e != nil {
				c.Warn("cannot read cmd from %s", c.RemoteAddr().String())
				c.Close()
				return
			}
			switch nM := m.(type) {
			case *msg.NewClient:
				c.ReplacePrefix("conn","ctrl")
				if token != (*nM).Token {
					c.Warn("wrong token taken from %s", c.RemoteAddr().String())
					c.Close()
					return
				}

				now := time.Now()
				client := &Client{
					ID:       strconv.Itoa(int(time.Now().Unix())) + strconv.Itoa(rand.Intn(10)),
					Conn:     c,
					Ports:    nM.Forwards,
					Proxies:  make(chan Conn, 10),
					LastPing: &now,
				}
				c.AddPrefix("client-" + client.ID)
				reg.Register(client.ID, client)
				c.Info("client registered")
				msg.WriteMsg(c, msg.ReNewClient{
					ID: client.ID,
				})
				msg.WriteMsg(c, msg.NewProxy{})
				go client.AcceptCmd()
				go client.StartListener()
			case *msg.NewProxy:
				c.ReplacePrefix("conn","proxy")
				oC, ok := reg[nM.ClientID]
				if !ok {
					util.Info("MultiEx client %s contains wrong client id", c.RemoteAddr().String())
					c.Close()
					break
				}
				oC.Proxies <- c
				c.AddPrefix("client-" + oC.ID)
				oC.Conn.Info("a new proxy connection added")
			}
		}()
	}
}
