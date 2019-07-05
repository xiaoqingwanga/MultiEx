package server

import (
	"MultiEx/util"
	"flag"
)

type options struct {
	clientPort string
	token      string
	logLevel   string
	logTo      string
}



// Main is server entry point.
func Main() {
	options := option()
	util.Initlog(options.logLevel, options.logTo)

	var clientRegistry ClientRegistry = make(map[string]*Client)

	// Listen for MultiEx client connections and handle request
	HandleClient(options.clientPort, options.token, clientRegistry)
}

func option() options {
	clientPort := flag.String("clientPort", ":8070", "the port listening for MultiEx client.")
	token := flag.String("token", "", "Token is the credential client should hold to connect this server.Server doesn't have token default.")
	logLevel := flag.String("logLevel", "INFO", "the log level of this program.")
	logTo := flag.String("logTo", "stdout", "the location where logs save. Empty value and stdout have special meaning")
	flag.Parse()
	return options{
		token:      *token,
		clientPort: *clientPort,
		logLevel:   *logLevel,
		logTo:      *logTo,
	}
}