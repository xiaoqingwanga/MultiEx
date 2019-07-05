package server

import (
	"MultiEx/msg"
	"MultiEx/util"
	"math/rand"
	"time"
)



// HandleClient accept client control connection,proxy connection
func HandleClient(token string, port string, reg ClientRegistry) {
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
				c.Warn("cannot read message from %s",c.RemoteAddr().String())
				c.Close()
				return
			}
			c.Info("successfully read message")
			switch nM := m.(type) {
			case *msg.NewClient:
				if token != (*nM).Token {
					c.Warn("wrong token taken from %s",c.RemoteAddr().String())
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
				reg.Register(client.ID, client)
				c.Info("client registered")
				msg.WriteMsg(c, msg.ReNewClient{
					ID: client.ID,
				})
				msg.WriteMsg(c,msg.NewProxy{})
				go client.AcceptCmd()
				go client.StartListener()
			case *msg.NewProxy:
				oC,ok := reg[nM.ClientID]
				if !ok{
					util.Info("MultiEx client %s contains wrong client id",c.RemoteAddr().String())
					c.Close()
					break
				}
				oC.Proxies <- c
				c.AddPrefix("client-"+oC.ID)
				oC.Conn.Info("a new proxy connection added")
			}
		}()
	}
}


