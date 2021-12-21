package config

import (
	"encoding/json"
	"errors"
	"os"
)

const (
	DARK_THEME  string = "dark"
	LIGHT_THEME        = "light"

	DEFAULT           = LIGHT_THEME
	DEFAULT_CODESTYLE = "monokai"
)

func IsChromaTheme(theme string) bool {
	themes := []string{
		"abap",
		"api",
		"algol_nu",
		"arduino",
		"autumn",
		"borland",
		"bw",
		"colorful",
		"dracula",
		"emacs",
		"friendly",
		"fruity",
		"github",
		"igor",
		"lovelace",
		"manni",
		"monokai",
		"monokailight",
		"murphy",
		"native",
		"paraiso_dark",
		"paraiso_light",
		"pastie",
		"perldoc",
		"pygments",
		"rainbow_dash",
		"rrt",
		"solarized_dark",
		"solarized_dark256",
		"solarized_light",
		"swapoff",
		"tango",
		"trac",
		"vim",
		"vs",
		"xcode",
	}

	for _, th := range themes {
		if th == theme {
			return true
		}
	}

	return false
}

type ServiceConfig struct {
	Theme          string `json:"theme"`
	CodeBlockTheme string `json:"codeblock"`
	Port           int    `json:"port"` // Defaults to 0 if not set.
}

func ReadConfigFromFile(configFilename string) (*ServiceConfig, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	absPathToFile := home + "/" + configFilename

	data, err := os.ReadFile(absPathToFile)
	if err != nil {
		// User most likely has not set config. Return default instead.
		return &ServiceConfig{
			Theme:          DEFAULT,
			CodeBlockTheme: DEFAULT_CODESTYLE,
		}, nil
	}

	// Read whatever json fields into config variable.
	var conf ServiceConfig
	if err := json.Unmarshal(data, &conf); err != nil {
		return nil, errors.New(`The Json config file is poorly formatted.
Please check your config file again.

An example config file would look as such:
{
	"theme": "dark",
	"codeblock": "monokai",
	"port": 3000
}

NOTE: the last line does not have a trailing comma.

` + err.Error())
	}

	if conf.Theme == "" {
		conf.Theme = DEFAULT
	}
	if conf.CodeBlockTheme == "" || !IsChromaTheme(conf.CodeBlockTheme) {
		conf.CodeBlockTheme = DEFAULT_CODESTYLE
	}

	return &conf, nil
}
