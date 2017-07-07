package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestZeroConfig(t *testing.T) {
	assert := assert.New(t)

	cfg := Config{}
	assert.Equal(cfg.data, "")
}

func TestConfig(t *testing.T) {
	assert := assert.New(t)

	cfg := Config{data: "/data"}
	assert.Equal(cfg.fqdn, "/data")
}
