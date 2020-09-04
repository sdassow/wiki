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
			fmt.Printf("Starting on %s...\n", cfg.bind)

			srv, err := NewServer(cfg)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			srv.ListenAndServe()
		},
	}

	// Setup command line arguments and link to config file properties
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file")
	rootCmd.PersistentFlags().StringVarP(&cfg.bind, "bind", "b", "0.0.0.0:8000", "[int]:<port> to bind to")
	rootCmd.PersistentFlags().StringVarP(&cfg.brand, "brand", "", "Wiki", "branding at top of each page")
	rootCmd.PersistentFlags().StringVarP(&cfg.data, "data", "", "./data", "path to data")
	rootCmd.PersistentFlags().StringVarP(&cfg.csrf.keyfile, "csrf-keyfile", "", "./csrf.key", "path to csrf key file")
	rootCmd.PersistentFlags().BoolVarP(&cfg.csrf.insecure, "csrf-insecure", "", true, "send csrf cookie over http")
	rootCmd.PersistentFlags().StringVarP(&cfg.git.url, "git-url", "", "", "url to git repository")
	rootCmd.PersistentFlags().BoolVarP(&cfg.git.push, "git-push", "", true, "push to git repository")

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	viper.BindPFlag("bind", rootCmd.PersistentFlags().Lookup("bind"))
	viper.BindPFlag("brand", rootCmd.PersistentFlags().Lookup("brand"))
	viper.BindPFlag("data", rootCmd.PersistentFlags().Lookup("data"))
	viper.BindPFlag("csrf-keyfile", rootCmd.PersistentFlags().Lookup("csrf-keyfile"))
	viper.BindPFlag("csrf-insecure", rootCmd.PersistentFlags().Lookup("csrf-insecure"))
	viper.BindPFlag("git-url", rootCmd.PersistentFlags().Lookup("git-url"))
	viper.BindPFlag("git-push", rootCmd.PersistentFlags().Lookup("git-push"))

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initConfig() {
	// Read in environment variables with prefix WIKI_
	viper.SetEnvPrefix("WIKI")
	viper.AutomaticEnv()

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
		if err := viper.ReadInConfig(); err != nil {
			fmt.Println("Can't read config:", err)
			os.Exit(1)
		}

	}

	cfg.bind = viper.GetString("bind")
	cfg.brand = viper.GetString("brand")
	cfg.data = viper.GetString("data")
	cfg.csrf.keyfile = viper.GetString("csrf-keyfile")
	cfg.csrf.insecure = viper.GetBool("csrf-insecure")
	cfg.git.url = viper.GetString("git-url")
	cfg.git.push = viper.GetBool("git-push")
}
