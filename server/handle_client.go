package server

import (
	"MultiEx/log"
	"MultiEx/msg"
	"strconv"
	"time"
)

// HandleClient accept client control connection、proxy connection
func HandleClient(port string, token string, reg *ClientRegistry) {
	// Listen
	l := listen(port, reg)

	// Get and Handle new connection
	for nc := range l.conns {
		go func(c Conn) {
			defer func() {
				if r := recover(); r != nil {
					c.Warn("%v", r)
				}
			}()

			m, e := msg.ReadMsg(c)
			if e != nil {
				c.Warn("cannot read msg from %s", c.RemoteAddr().String())
				c.Close()
				// Maybe denial of service attack
				return
			}
			switch nM := m.(type) {
			case *msg.NewClient:
				c.ReplacePrefix("conn", "ctrl")
				if token != (*nM).Token {
					c.Warn("wrong token taken from %s", c.RemoteAddr().String())
					c.Close()
					return
				}


				clientCounter.Inc()
				now := time.Now()
				client := &Client{
					ID:       strconv.Itoa(int(clientCounter.Get())),
					Conn:     c,
					Ports:    nM.Forwards,
					Proxies:  make(chan Conn, 30),
					LastPing: &now,
				}
				c.AddPrefix("client-" + client.ID)
				reg.Register(client.ID, client)
				c.Info("client registered")
				msg.WriteMsg(c, msg.ReNewClient{
					ID: client.ID,
				})
				msg.WriteMsg(c, msg.NewProxy{})
				go client.AcceptCmd(reg)
				go client.StartListener()
			case *msg.NewProxy:
				c.ReplacePrefix("conn", "proxy")
				oC, ok := (*reg)[nM.ClientID]
				if !ok {
					log.Info("MultiEx client %s contains wrong client id", c.RemoteAddr().String())
					c.Close()
					break
				}
				oC.Proxies <- c
				c.AddPrefix("client-" + oC.ID)
				oC.Conn.Info("a new proxy connection added")
			}
		}(nc)
	}
}
