package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/norkeel/terraform-provider-norkeel/internal/provider"
)

// version is set by GoReleaser (-X main.version=…); "dev" for local builds.
var version = "dev"

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "run with support for debuggers (delve)")
	flag.Parse()

	err := providerserver.Serve(context.Background(), provider.New(version), providerserver.ServeOpts{
		Address: "registry.terraform.io/norkeel/norkeel",
		Debug:   debug,
	})
	if err != nil {
		log.Fatal(err)
	}
}
