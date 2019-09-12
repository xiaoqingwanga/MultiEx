package server

import (
	"MultiEx/log"
	"MultiEx/msg"
	"MultiEx/util"
	"strconv"
	"sync"
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

			m, e,_ := msg.ReadMsg(c)
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

				now := time.Now()
				var inUsePort util.Count
				client := &Client{
					Conn:      c,
					Ports:     nM.Forwards,
					InUsePort: inUsePort,
					Proxies:   make(chan Conn, 30),
					LastPing:  &now,
				}

				var wg sync.WaitGroup
				client.StartListener(&wg)
				wg.Wait()
				if int(client.InUsePort.Get()) == len(client.Ports) {
					client.Conn.Warn("all port %v not available.this client not work,abort..", client.Ports)
					client.Close()
					return
				}
				clientCounter.Inc()
				client.ID = strconv.Itoa(int(clientCounter.Get()))
				c.AddPrefix("client-" + client.ID)
				reg.Register(client.ID, client)
				c.Info("client registered")
				msg.WriteMsg(c, msg.ReNewClient{
					ID: client.ID,
				})
				msg.WriteMsg(c, msg.NewProxy{})
				go client.AcceptCmd(reg)
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
