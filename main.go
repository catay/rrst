package main

import (
	"fmt"
	"github.com/catay/rrst/cmd"
	"os"
)

func main() {
	c := cmd.New()

	if err := c.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}
}
