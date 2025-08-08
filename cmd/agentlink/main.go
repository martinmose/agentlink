package main

import (
	"os"

	"github.com/martinmose/agentlink/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}