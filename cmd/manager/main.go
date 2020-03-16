package main

import (
	"github.com/keikoproj/manager/cmd/manager/commands"
	"github.com/prometheus/common/log"
)

func main() {
	err := commands.NewCommand().Execute()
	if err != nil {
		log.Fatal(err.Error())
	}
}
