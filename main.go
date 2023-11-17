package main

import (
	"context"
	"flag"
	"log"

	"github.com/fivetran/terraform-provider-fivetran/fivetran"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tf5server"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
)

func main() {

	ctx := context.Background()

	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	providers := []func() tfprotov5.ProviderServer{
		providerserver.NewProtocol5(framework.FivetranProvider()), // Example terraform-plugin-framework provider
		fivetran.Provider().GRPCProvider,                          // Example terraform-plugin-sdk provider
	}

	muxServer, err := tf5muxserver.NewMuxServer(ctx, providers...)

	if err != nil {
		log.Fatal(err)
	}

	var serveOpts []tf5server.ServeOpt

	if debug {
		serveOpts = append(serveOpts, tf5server.WithManagedDebug())
	}

	err = tf5server.Serve(
		"registry.terraform.io/<namespace>/<provider_name>",
		muxServer.ProviderServer,
		serveOpts...,
	)

	if err != nil {
		log.Fatal(err)
	}
	// plugin.Serve(&plugin.ServeOpts{
	// 	ProviderFunc: func() *schema.Provider {
	// 		return fivetran.Provider()
	// 	},
	// })
}
