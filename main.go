package main

import (
	"context"
	"flag"
	"log"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

//go:generate go run utils/generate_connector_config.go

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/providers/fivetran/fivetran",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), framework.FivetranProvider, opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
