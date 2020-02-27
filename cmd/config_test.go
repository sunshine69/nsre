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

func TestTernary(t *testing.T) {
	l0 := []string {"1", "2"}
	length := len(l0)
	fmt.Println(length)
	var o string
	var test bool = length == 3
	// if test {
	// 	o = l0[2]
	// } else { o = l0[0] }
	o = Ternary(test, l0[2], l0[0] ).(string)
	fmt.Println(o)
}