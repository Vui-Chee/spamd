package config

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/alecthomas/chroma/v2/styles"
)

func TestIsChromaTheme(t *testing.T) {
	var got bool

	// Invalid theme should return false.
	got = IsChromaTheme("asafd")
	if got != false {
		t.Error("got: true; want: false")
	}

	for _, validTheme := range styles.Names() {
		got = IsChromaTheme(validTheme)
		if got != true {
			t.Error("got: false; want: true")
		}
	}
}

func TestSetValidTheme(t *testing.T) {
	testConfig := ServiceConfig{
		Theme: "light",
	}

	testConfig.SetTheme("dark")
	if testConfig.Theme != "dark" {
		t.Errorf("got: %s; want: dark\n", testConfig.Theme)
	}

	testConfig.SetTheme("light")
	if testConfig.Theme != "light" {
		t.Errorf("got: %s; want: light\n", testConfig.Theme)
	}
}

func TestSetInvalidTheme(t *testing.T) {
	testConfig := ServiceConfig{
		Theme: "light",
	}

	testConfig.SetTheme("some non-existent theme")
	if testConfig.Theme != "light" {
		t.Errorf("got: %s; want: light\n", testConfig.Theme)
	}
}

func TestSetEmptyCodeTheme(t *testing.T) {
	testConfig := ServiceConfig{
		CodeBlockTheme: "vim",
	}
	err := testConfig.SetCodeBlockTheme("")
	if err != nil {
		// Config should end in default code theme.
		t.Error("Empty code theme should not return error.")
	}
	if testConfig.CodeBlockTheme != "vim" {
		t.Errorf("got %s; want vim", testConfig.CodeBlockTheme)
	}
}

func TestErrorOnInvalidCodeTheme(t *testing.T) {
	testConfig := ServiceConfig{
		CodeBlockTheme: "vim",
	}

	// Ensure invalidTheme does not exist in set
	// of valid themes.
	invalidTheme := "asdf"
	for _, th := range styles.Names() {
		if th == invalidTheme {
			t.Errorf("%s should not included in themes.", invalidTheme)
		}
	}

	err := testConfig.SetCodeBlockTheme(invalidTheme)
	if err == nil {
		t.Error("got <nil>; want error")
	}
	if testConfig.CodeBlockTheme != "vim" {
		t.Errorf("got %s; want vim", testConfig.CodeBlockTheme)
	}
}

func TestSetValidCodeTheme(t *testing.T) {
	testConfig := ServiceConfig{
		CodeBlockTheme: "vim",
	}

	err := testConfig.SetCodeBlockTheme("xcode")
	if err != nil {
		t.Errorf("Setting valid code theme should not result in error. Got %s", err)
	}
	if testConfig.CodeBlockTheme != "xcode" {
		t.Errorf("got %s; want xcode", testConfig.CodeBlockTheme)
	}
}

func TestConstructConfigFromFile(t *testing.T) {
	home, _ := os.UserHomeDir()
	file, _ := ioutil.TempFile(home, ".*")
	file.WriteString("{\"theme\":\"dark\",\"port\":1234,\"codeblock\":\"vim\"}")
	defer os.Remove(file.Name())

	configFilename := path.Base(file.Name()[1:])
	conf, _ := ReadConfigFromFile(configFilename)

	if conf.Theme != "dark" {
		t.Errorf("got: %s; want: dark", conf.Theme)
	}
	if conf.Port != 1234 {
		t.Errorf("got: %d; want: 1234", conf.Port)
	}
	if conf.CodeBlockTheme != "vim" {
		t.Errorf("got: %s; want: vim", conf.CodeBlockTheme)
	}
}

func TestConstructConfigFromDefaults(t *testing.T) {
	home, _ := os.UserHomeDir()
	file, _ := ioutil.TempFile(home, ".*")
	file.WriteString("{\"port\":1234}")
	defer os.Remove(file.Name())

	configFilename := path.Base(file.Name()[1:])
	conf, _ := ReadConfigFromFile(configFilename)

	if conf.Theme != DEFAULT {
		t.Errorf("got: %s; want: %s", conf.Theme, DEFAULT)
	}
	if conf.Port != 1234 {
		t.Errorf("got: %d; want: 1234", conf.Port)
	}
	if conf.CodeBlockTheme != DEFAULT_CODESTYLE {
		t.Errorf("got: %s; want: %s", conf.CodeBlockTheme, DEFAULT_CODESTYLE)
	}
}

func TestSetDefaultOnAbsentConfigFile(t *testing.T) {
	home, _ := os.UserHomeDir()
	configFilename := "nosuchfile"

	_, err := os.Lstat(home + "/" + configFilename)
	if err == nil {
		t.Error("Config file should be absent from ROOT directory.")
	}

	conf, err := ReadConfigFromFile(configFilename)
	if err != nil {
		t.Error("Should not return error if config file is absent.")
	}
	if conf.Theme != DEFAULT {
		t.Errorf("got: %s; want: %s", conf.Theme, DEFAULT)
	}
	if conf.CodeBlockTheme != DEFAULT_CODESTYLE {
		t.Errorf("got: %s; want: %s", conf.CodeBlockTheme, DEFAULT_CODESTYLE)
	}
}

func TestErrorOnInvalidConfigFile(t *testing.T) {
	// File exists but got invalid config
	home, _ := os.UserHomeDir()
	file, _ := ioutil.TempFile(home, ".*")
	// Config with trailing comma in last line
	file.WriteString("{\"theme\":\"dark\",\"port\":1234,}")
	defer os.Remove(file.Name())

	configFilename := path.Base(file.Name()[1:])
	conf, err := ReadConfigFromFile(configFilename)
	if conf != nil {
		t.Error("Should return <nil> on invalid config file.")
	}
	if err == nil {
		t.Error("Should return error on invalid config file.")
	}
}
