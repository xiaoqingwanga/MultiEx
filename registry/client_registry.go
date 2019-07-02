package registry

import "MultiEx/server"

// ClientRegistry is a place storing clients.
type ClientRegistry map[string]*server.Client

// Register register client
func (registry *ClientRegistry) Register(id string, client *server.Client) (oClient *server.Client) {
	oClient, ok := (*registry)[id]
	if ok {
		return
	}
	(*registry)[id] = client
	return
}
