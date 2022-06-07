package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/everadaptive/mindlights/eeg/neurosky"
	"github.com/everadaptive/mindlights/service"
	"go.uber.org/zap"

	"github.com/vmware/transport-go/plank/pkg/server"
	"github.com/vmware/transport-go/plank/utils"
)

var (
	log *zap.SugaredLogger
)

// configure flags
func main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync() // flushes buffer, if any
	log = logger.Sugar()

	neurosky, err := neurosky.NewNeurosky("98:D3:31:80:7B:3D", "HEADSET-03")
	if err != nil {
		log.Fatal(err)
	}
	events := neurosky.Start()

	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, os.Interrupt)

	go func() {
		sig := <-signalChan
		fmt.Printf("Got %s signal. Aborting...\n", sig)
		neurosky.Close()
	}()

	serverConfig, err := server.CreateServerConfig()
	if err != nil {
		utils.Log.Fatalln(err)
	}
	platformServer := server.NewPlatformServer(serverConfig)
	if err = platformServer.RegisterService(service.NewNeuroskyService(events), service.NeuroskyServiceChan); err != nil {
		utils.Log.Fatalln(err)
	}
	syschan := make(chan os.Signal, 1)
	platformServer.StartServer(syschan)
}
