package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/everadaptive/mindlights/udmx"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	// Used for flags.
	cfgFile     string
	displayType string
	displaySize int

	envPrefix = "MINDLIGHTS"

	rootCmd = &cobra.Command{
		Use:   "mindlights",
		Short: "Use data from a Neurosky to control things",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// You can bind cobra and viper in a few locations, but PersistencePreRunE on the root command works well
			return initializeConfig(cmd)
		},
		Run: func(cmd *cobra.Command, args []string) {
			var (
				// disp      display.ColorDisplay
				// palette   []colorful.Color
				dmxDevice udmx.DmxDevice
			)
			logger, _ := zap.NewDevelopment()
			defer logger.Sync() // flushes buffer, if any
			log = logger.Sugar()

			if displayType == "udmx" {
				dmxDevice = &udmx.UDmxDevice{}

				dmxDevice.Open()
				defer dmxDevice.Close()

				// disp = display.NewUDmxDisplay(displaySize, dmxDevice)
			} else if displayType == "serialdmx" {
				dmxDevice = &udmx.SerialDMXDevice{}

				dmxDevice.Open()
				defer dmxDevice.Close()

				// disp = display.NewSerialDmxDisplay(displaySize, dmxDevice)
			} else if displayType == "dummy" {
				// disp = display.NewDummyDisplay()
			}

			// palette = controller.CustomPalette6()

			signalChan := make(chan os.Signal)
			signal.Notify(signalChan, os.Interrupt)

			go func() {
				sig := <-signalChan
				fmt.Printf("Got %s signal. Aborting...\n", sig)
			}()

			dmxDevice.SetChannelColor(1, 255)
			dmxDevice.Render()
		},
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "./mindlights.yaml", "config file")
	rootCmd.PersistentFlags().StringVar(&displayType, "display", "dummy", "output display 'dummy', 'udmx', 'serialdmx'")
	rootCmd.PersistentFlags().IntVar(&displaySize, "display-size", 8, "output display size")
}

func initializeConfig(cmd *cobra.Command) error {
	v := viper.New()

	if cfgFile != "" {
		// Use config file from the flag.
		v.SetConfigFile(cfgFile)
	} else {
		// Find current working directory.
		cwd, err := os.Getwd()
		cobra.CheckErr(err)

		// Search config in current working directory with name ".yamle" (without extension).
		v.AddConfigPath(cwd)
		v.SetConfigType("yaml")
		v.SetConfigName("mindlights")
	}

	if err := v.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", v.ConfigFileUsed())
	}

	// When we bind flags to environment variables expect that the
	// environment variables are prefixed, e.g. a flag like --number
	// binds to an environment variable STING_NUMBER. This helps
	// avoid conflicts.
	v.SetEnvPrefix(envPrefix)

	// Bind to environment variables
	// Works great for simple config names, but needs help for names
	// like --favorite-color which we fix in the bindFlags function
	v.AutomaticEnv()

	// Bind the current command's flags to viper
	bindFlags(cmd, v)

	return nil
}

// Bind each cobra flag to its associated viper configuration (config file and environment variable)
func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Environment variables can't have dashes in them, so bind them to their equivalent
		// keys with underscores, e.g. --favorite-color to STING_FAVORITE_COLOR
		if strings.Contains(f.Name, "-") {
			envVarSuffix := strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
			v.BindEnv(f.Name, fmt.Sprintf("%s_%s", envPrefix, envVarSuffix))
		}

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}
