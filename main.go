package main

import (
	"github.com/hashicorp/terraform/plugin"
	"log"
	"os"
	"terraform-provider-signalform/signalform"
)

func main() {
	logFile, err := os.OpenFile("signalform.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	// direct all log messages to signalform.log
	log.SetOutput(logFile)

	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: signalform.Provider,
	})
}
