package cmd

import (
	"os"
	"fmt"
	"testing"
)

func TestConfigKey(t *testing.T) {
	defaultConfig :=  fmt.Sprintf("%s/.nsre-dev1.yaml", os.Getenv("HOME"))
	LoadConfig(defaultConfig)
	fmt.Printf("DB PATH: %s\n", Config.Logdbpath)
	p0 := GetConfigSave("twilio_sid", "1qa")
	fmt.Printf("DEBUG %s\n", p0)
	DeleteConfig("twilio_sid")
}