package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/everadaptive/mindlights/eeg/neurosky"
	"github.com/everadaptive/mindlights/service"
	"github.com/spf13/cobra"
	"github.com/vmware/transport-go/bus"
	"github.com/vmware/transport-go/plank/pkg/server"
	"github.com/vmware/transport-go/plank/utils"
	"go.uber.org/zap"
)

var (
	streamSpaPath string
	streamPort    int
	streamHost    string
)

var streamCmd = &cobra.Command{
	Use:   "stream",
	Short: "Stream EEG data to the web UI over WebSocket",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initializeConfig(cmd)
	},
	Run: func(cmd *cobra.Command, args []string) {
		logger, _ := zap.NewDevelopment()
		defer logger.Sync()
		log = logger.Sugar()

		if len(eegHeadsets) == 0 {
			log.Fatal("no eeg_headsets defined in config")
		}
		h := eegHeadsets[0]

		ns, err := neurosky.NewNeurosky(h.BluetoothAddress, h.Name, log.Named(h.Name))
		if err != nil {
			log.Fatal(err)
		}
		events := ns.Start()

		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt)
		go func() {
			sig := <-signalChan
			fmt.Printf("Got %s signal. Aborting...\n", sig)
			ns.Close()
		}()

		cwd, _ := os.Getwd()
		serverConfig := &server.PlatformServerConfig{
			Host:    streamHost,
			Port:    streamPort,
			RootDir: cwd,
			FabricConfig: &server.FabricBrokerConfig{
				FabricEndpoint: "/ws",
				EndpointConfig: &bus.EndpointConfig{
					TopicPrefix:           "/topic",
					AppRequestPrefix:      "/pub",
					AppRequestQueuePrefix: "/pub/queue",
					UserQueuePrefix:       "/queue",
				},
			},
			LogConfig: &utils.LogConfig{
				OutputLog:     "stdout",
				AccessLog:     "stdout",
				ErrorLog:      "stderr",
				FormatOptions: &utils.LogFormatOption{},
			},
			ShutdownTimeout:   5 * time.Minute,
			RestBridgeTimeout: 1 * time.Minute,
			NoBanner:          true,
		}

		if streamSpaPath != "" {
			spaConfig, err := server.NewSpaConfig(streamSpaPath)
			if err != nil {
				log.Fatal(err)
			}
			serverConfig.SpaConfig = spaConfig
		}

		platformServer := server.NewPlatformServer(serverConfig)
		if err = platformServer.RegisterService(service.NewNeuroskyService(events, h.Name, oscHost, oscPort), service.NeuroskyServiceChan); err != nil {
			utils.Log.Fatalln(err)
		}
		syschan := make(chan os.Signal, 1)
		platformServer.StartServer(syschan)
	},
}

func init() {
	streamCmd.Flags().StringVar(&streamSpaPath, "spa-path", "", "path to built web UI static files to serve (e.g. ./app)")
	streamCmd.Flags().IntVar(&streamPort, "port", 30080, "port to serve on")
	streamCmd.Flags().StringVar(&streamHost, "host", "localhost", "hostname to bind")
	rootCmd.AddCommand(streamCmd)
}
