package main

import (
	"github.com/catay/rrst/cmd"
	"log"
)

func main() {
	c := cmd.New()

	if err := c.Run(); err != nil {
		log.Fatal(err)
	}
}
