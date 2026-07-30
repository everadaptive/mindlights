package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/everadaptive/mindlights/eeg/neurosky"
	"github.com/everadaptive/mindlights/service"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/vmware/transport-go/plank/pkg/server"
	"github.com/vmware/transport-go/plank/utils"
)

var (
	cfgFile     string
	oscHost     string
	oscPort     int
	eegHeadsets []eegHeadsetConfig
	log         *zap.SugaredLogger

	envPrefix = "MINDLIGHTS"

	rootCmd = &cobra.Command{
		Use:   "stream",
		Short: "Stream EEG data to the web UI",
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

			serverConfig, err := server.CreateServerConfig()
			if err != nil {
				utils.Log.Fatalln(err)
			}
			platformServer := server.NewPlatformServer(serverConfig)
			if err = platformServer.RegisterService(service.NewNeuroskyService(events, h.Name, oscHost, oscPort), service.NeuroskyServiceChan); err != nil {
				utils.Log.Fatalln(err)
			}
			syschan := make(chan os.Signal, 1)
			platformServer.StartServer(syschan)
		},
	}
)

func main() {
	rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "./mindlights.yaml", "config file")
	rootCmd.PersistentFlags().StringVar(&oscHost, "osc-host", "", "OSC destination host (disabled if empty)")
	rootCmd.PersistentFlags().IntVar(&oscPort, "osc-port", 9000, "OSC destination port")
}

func initializeConfig(cmd *cobra.Command) error {
	v := viper.New()

	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {
		cwd, err := os.Getwd()
		cobra.CheckErr(err)
		v.AddConfigPath(cwd)
		v.SetConfigType("yaml")
		v.SetConfigName("mindlights")
	}

	if err := v.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", v.ConfigFileUsed())
	}

	v.SetEnvPrefix(envPrefix)
	v.AutomaticEnv()
	bindFlags(cmd, v)
	v.UnmarshalKey("eeg_headsets", &eegHeadsets)

	return nil
}

func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if strings.Contains(f.Name, "-") {
			envVarSuffix := strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
			v.BindEnv(f.Name, fmt.Sprintf("%s_%s", envPrefix, envVarSuffix))
		}
		if !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}
