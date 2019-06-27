package server

import (
	"MultiEx/util"
	"flag"
)

type options struct {
	clientPort string
	logLevel   string
	logTo      string
}

func getOptions() options {
	clientPort := flag.String("clientPort", ":8070", "the port listening for MultiEx client.")
	logLevel := flag.String("logLevel", "INFO", "the log level of this program.")
	logTo := flag.String("logTo", "stdout", "the location where logs save. Empty value and stdout have special meaning")
	return options{
		clientPort: *clientPort,
		logLevel:   *logLevel,
		logTo:      *logTo,
	}
}

// Main is server entry point.
func Main() {
	options := getOptions()
	util.Initlog(options.logLevel, options.logTo)

	// Listen for MultiEx client connections and handle request
	HandleClient(options.clientPort)
}
