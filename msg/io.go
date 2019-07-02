package msg

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
)

// ReadMsg read bytes from reader and convert them to a message.
func ReadMsg(r io.Reader) (m Message, e error) {
	// Read size of message.
	var size int16
	e = binary.Read(r, binary.LittleEndian, &size)
	if e != nil {

		return
	}
	// Read message bytes and convert to json object.
	bytes := make([]byte, size)
	rSize, e := r.Read(bytes)
	if e != nil {
		return
	}
	if int16(rSize) != size {
		e = fmt.Errorf("read size is not equal original size")
		return
	}
	var pkg pack
	e = json.Unmarshal(bytes, &pkg)
	if e != nil {
		return
	}

	switch pkg.Typ {
	case "NewClient":
		m = &NewClient{}
	case "ReNewClient":
		m = &ReNewClient{}
	case "NewProxy":
		m = &NewProxy{}
	case "CloseProxy":
		m = &CloseProxy{}
	case "ActivateProxy":
		m = &ActivateProxy{}
	case "Ping":
		m = &Ping{}
	case "Pong":
		m = &Pong{}
	case "PortInUse":
		m = &PortInUse{}
	case "CloseCtrl":
		m = &CloseCtrl{}
	case "GResponse":
		m = &GResponse{}
	default:
		e = fmt.Errorf("cannot parse connection type")
		return
	}
	e = json.Unmarshal(pkg.Msg, m)
	return
}

// WriteMsg write message to writer.
func WriteMsg(w io.Writer, msg Message) (e error) {

	typ := reflect.TypeOf(msg).Name()

	if e != nil {
		return
	}
	pBytes, e := json.Marshal(struct {
		Typ string
		Msg interface{}
	}{
		Typ: typ,
		Msg: msg,
	})
	if e != nil {
		return
	}
	pLen := int16(len(pBytes))
	e = binary.Write(w, binary.LittleEndian, pLen)
	if e != nil {
		return
	}
	len, e := w.Write(pBytes)
	if e != nil {
		return
	}
	if int16(len) != pLen {
		e = fmt.Errorf("write package to writer failed")
	}
	return
}
