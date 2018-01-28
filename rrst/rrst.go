package rrst

import (
	"fmt"
	"github.com/catay/rrst/config"
)

type app struct {
	config *config.ConfigData
}

func New(configFile string) (a *app, err error) {
	a = new(app)
	a.config, err = config.New(configFile)
	if err != nil {
		return nil, err
	}

	return
}

func (self *app) Print() {
	for _, r := range self.config.Repos {
		fmt.Println("*", r.Name)
	}
}

func (self *app) Sync() (err error) {
	fmt.Println("Not implemented yet!")
	return
}
