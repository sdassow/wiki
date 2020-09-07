package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"strings"
)

var (
	cfg     Config
	cfgFile string
)

func main() {

	// The sole command
	var rootCmd = &cobra.Command{
		Use: "wiking",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Listening on %s (%s/%s)...\n", cfg.listen.address, cfg.listen.network, cfg.listen.protocol)

			srv, err := NewServer(cfg)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			srv.ListenAndServe()
		},
	}

	cobra.OnInitialize(initConfig)
	rootCmd.Flags().StringVarP(&cfgFile, "config", "c", "", "path to config file")

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	viper.SetDefault("listen-address", ":8000")
	viper.SetDefault("listen-network", "tcp")
	viper.SetDefault("listen-protocol", "http")
	viper.SetDefault("brand", "Wiki")
	viper.SetDefault("csrf-keyfile", "./csrf.key")
	viper.SetDefault("csrf-insecure", false)
	viper.SetDefault("data", "./data")
	viper.SetDefault("git-push", true)
	viper.SetDefault("git-url", "")
	viper.SetDefault("indexdir", "./riot-index")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initConfig() {
	// Read in environment variables with prefix WIKING_
	viper.SetEnvPrefix("WIKING")
	viper.AutomaticEnv()

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
		if err := viper.ReadInConfig(); err != nil {
			fmt.Println("Can't read config:", err)
			os.Exit(1)
		}

	}

	cfg.listen.address = viper.GetString("listen-address")
	cfg.listen.network = viper.GetString("listen-network")
	cfg.listen.protocol = viper.GetString("listen-protocol")
	cfg.brand = viper.GetString("brand")
	cfg.data = viper.GetString("data")
	cfg.indexdir = viper.GetString("indexdir")
	cfg.csrf.keyfile = viper.GetString("csrf-keyfile")
	cfg.csrf.insecure = viper.GetBool("csrf-insecure")
	cfg.git.url = viper.GetString("git-url")
	cfg.git.push = viper.GetBool("git-push")
}
