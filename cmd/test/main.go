package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/erwanlbp/trading-bot/pkg/config"
	"github.com/erwanlbp/trading-bot/pkg/util"
	"go.uber.org/zap"
)

func main() {
	cancelChan := make(chan os.Signal, 1)
	// catch SIGETRM or SIGINTERRUPT
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)

	conf := config.Init(context.Background())

	fmt.Println("current conf")
	util.DebugPrintYaml(conf.ConfigFile)

	fmt.Println("Enabled coins")
	x, err := conf.Repository.GetEnabledCoins()
	if err != nil {
		fmt.Println("failed getting enabled coins", err)
	} else {
		util.DebugPrintJson(x)
	}

	// Wait for ctrl+c to reload conf
	fmt.Println("Waiting for ctrl+c to reload conf")
	<-cancelChan

	// Reload conf
	if err := conf.ReloadConfigFile(context.TODO()); err != nil {
		conf.Logger.Fatal("failed to reload conf", zap.Error(err))
	}

	fmt.Println("new conf")
	util.DebugPrintYaml(conf.ConfigFile)

	fmt.Println("jump finder process conf")
	util.DebugPrintYaml(conf.ProcessJumpFinder.ConfigFile)

	fmt.Println("New enabled coins")
	x, err = conf.Repository.GetEnabledCoins()
	if err != nil {
		fmt.Println("failed getting new enabled coins", err)
	} else {
		util.DebugPrintJson(x)
	}
}
