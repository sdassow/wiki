package main

import (
	"github.com/namsral/flag"
)

var (
	cfg Config
)

func main() {
	var (
		config string
		data   string
		bind   string
	)

	flag.StringVar(&config, "config", "", "config file")
	flag.StringVar(&data, "data", "./data", "path to data")
	flag.StringVar(&bind, "bind", "0.0.0.0:8000", "[int]:<port> to bind to")
	flag.Parse()

	// TODO: Abstract the Config and Handlers better
	cfg.data = data

	NewServer(bind, cfg).ListenAndServe()
}
