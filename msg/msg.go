package msg

import (
	"encoding/json"
)

// Message root; 'Dependency inversion principle'.
type Message interface {
}

type pack struct {
	Typ string
	Msg json.RawMessage
}

// NewClient request establish a new control connection.
type NewClient struct {
	Token string
	// Ports forwarded
	Forwards []string
}

// Response NewClient request
type ReNewClient struct {
	ID string
}

type CloseCtrl struct {
}

type Ping struct {
}

type Pong struct {
}

type PortInUse struct {
	Port string
}

// NewProxy request establish a new proxy connection.
type NewProxy struct {
	ClientID string
}

type CloseProxy struct {
}

type ActivateProxy struct {
}

type ClientNotExist struct {

}


// ForwardInfo tell MultiEx client which port public request
type ForwardInfo struct {
	Port string
}

// GResponse is general response, should not use
type GResponse struct {
	Msg string
}

