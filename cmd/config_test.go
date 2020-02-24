package cmd

import (
	"os"
	"fmt"
	"testing"
)

func TestConfigKey(t *testing.T) {
	defaultConfig :=  fmt.Sprintf("%s/.nsre-dev1.yaml", os.Getenv("HOME"))
	LoadConfig(defaultConfig)
	p0 := 
}