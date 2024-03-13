package main

import (
	"github.com/engpetarmarinov/gotama/internal/cli/cmd"
	"github.com/engpetarmarinov/gotama/internal/logger"
)

func main() {
	logger.Init(logger.NewConfigOpt().WithLevel(logger.DEBUG))
	cmd.Run()
}
