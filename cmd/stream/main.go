package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

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
	replayFile  string
	replaySpeed float64
	replayStart int
	replayLimit int
	serverHost  string
	serverPort  int
	spaPath     string
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

			var events chan neurosky.MindflexEvent

			if replayFile != "" {
				var err error
				events, err = replayCSV(replayFile, h.Name, replaySpeed, time.Duration(replayStart)*time.Second, time.Duration(replayLimit)*time.Second, log.Named("replay"))
				if err != nil {
					log.Fatal(err)
				}
			} else {
				ns, err := neurosky.NewNeurosky(h.BluetoothAddress, h.Name, log.Named(h.Name))
				if err != nil {
					log.Fatal(err)
				}
				events = ns.Start()

				signalChan := make(chan os.Signal, 1)
				signal.Notify(signalChan, os.Interrupt)
				go func() {
					sig := <-signalChan
					fmt.Printf("Got %s signal. Aborting...\n", sig)
					ns.Close()
				}()
			}

			// plank's CreateServerConfig re-parses os.Args with pflag.ExitOnError,
			// which would exit on our cobra flags; hide them to get a default config.
			savedArgs := os.Args
			os.Args = savedArgs[:1]
			serverConfig, err := server.CreateServerConfig()
			os.Args = savedArgs
			if err != nil {
				utils.Log.Fatalln(err)
			}
			serverConfig.Host = serverHost
			serverConfig.Port = serverPort
			if spaPath != "" {
				serverConfig.SpaConfig, err = server.NewSpaConfig(spaPath)
				if err != nil {
					utils.Log.Fatalln(err)
				}
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
	rootCmd.PersistentFlags().StringVar(&replayFile, "replay", "", "replay a CSV log instead of connecting a headset")
	rootCmd.PersistentFlags().Float64Var(&replaySpeed, "replay-speed", 1.0, "replay speed multiplier")
	rootCmd.PersistentFlags().IntVar(&replayStart, "replay-start", 0, "skip the first N seconds of the replay")
	rootCmd.PersistentFlags().IntVar(&replayLimit, "replay-limit", 0, "limit replay to the first N seconds (0 = all)")
	rootCmd.PersistentFlags().BoolVar(&legacyBandShift, "legacy-band-shift", true, "CSV was recorded before the EEG_POWER parser offset fix; shift band columns to their true names")
	rootCmd.PersistentFlags().StringVar(&serverHost, "host", "localhost", "host where the web UI is served")
	rootCmd.PersistentFlags().IntVar(&serverPort, "port", 30080, "port where the web UI is served")
	rootCmd.PersistentFlags().StringVar(&spaPath, "spa-path", "", `path to a built SPA to serve (URI derived from leaf dir; override with "path:/uri")`)
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
