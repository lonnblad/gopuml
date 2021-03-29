package main

import (
	"os"

	"github.com/lonnblad/gopuml/cmd/gopuml/internal"
)

var version string = "dev"

func main() {
	rootCmd := internal.CreateRootCmd()
	buildCmd := internal.CreateBuildCmd()
	serveCmd := internal.CreateServeCmd()
	versionCmd := internal.CreateVersionCmd(version)

	rootCmd.AddCommand(&buildCmd, &serveCmd, &versionCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}
