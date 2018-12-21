package main

import (
	"fmt"
	"github.com/catay/rrst/cmd/rrstd/cmd"
	"os"
)

func main() {
	c := cmd.NewCli()

	if err := c.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
