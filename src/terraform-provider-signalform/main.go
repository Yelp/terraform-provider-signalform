package main

import (
	"github.com/hashicorp/terraform/plugin"
	"terraform-provider-signalform/signalform"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: signalform.Provider,
	})
}
