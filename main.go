package main

import (
	"github.com/docker/machine/libmachine/drivers/plugin"
	"github.com/phoenixnap/docker-machine-driver-pnap/pkg/drivers/pnap"
)

func main() {
	plugin.RegisterDriver(pnap.NewDriver())
}
