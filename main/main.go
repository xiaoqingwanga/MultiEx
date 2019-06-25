package main

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
	clientPort := flag.String("clientPort", ":8070", "the port listening for client.")
	logLevel := flag.String("logLevel", "INFO", "the log level of this program.")
	logTo := flag.String("logTo", "stdout", "the location where logs save. Empty value and stdout have special meaning")
	return options{
		clientPort: *clientPort,
		logLevel:   *logLevel,
		logTo:      *logTo,
	}
}

// Main is server entry point.
func main() {
	util.Initlog("info", "stdout")
	pf := util.NewPrefixLogger("abc", "def")
	util.Info("hello")
	pf.Info("hello")
}
