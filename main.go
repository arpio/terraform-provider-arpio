package main

import (
	"github.com/arpio/terraform-provider-arpio/arpio"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

// Version is set through ldflags.
var Version = "0.0"

// Commit is set through ldflags.
var Commit = "<unknown>"

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() *schema.Provider {
			return arpio.Provider()
		},
	})
}
