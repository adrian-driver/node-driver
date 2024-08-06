package main

import (
	"github.com/docker/machine/libmachine/drivers/plugin"
	"github.com/adrian-driver/node-driver/pkg/drivers/poc"
)

func main() {
	plugin.RegisterDriver(poc.NewDriver())
}
