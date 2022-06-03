package main

import (
	"os"

	"github.com/everadaptive/mindlights/service"

	"github.com/vmware/transport-go/plank/pkg/server"
	"github.com/vmware/transport-go/plank/utils"
)

// configure flags
func main() {
	serverConfig, err := server.CreateServerConfig()
	if err != nil {
		utils.Log.Fatalln(err)
	}
	platformServer := server.NewPlatformServer(serverConfig)
	if err = platformServer.RegisterService(service.NewNeuroskyService(), service.NeuroskyServiceChan); err != nil {
		utils.Log.Fatalln(err)
	}
	syschan := make(chan os.Signal, 1)
	platformServer.StartServer(syschan)
}
