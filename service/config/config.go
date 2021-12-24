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

	invalid_config_error = `The Json config file is poorly formatted.
Please check your config file again.

An example config file would look as such:
{
	"theme": "dark",
	"codeblock": "monokai",
	"port": 3000
}

NOTE: the last line does not have a trailing comma.

`
)

var themes = []string{
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
	"paraiso-dark",
	"paraiso-light",
	"pastie",
	"perldoc",
	"pygments",
	"rainbow_dash",
	"rrt",
	"solarized-dark",
	"solarized-dark256",
	"solarized-light",
	"swapoff",
	"tango",
	"trac",
	"vim",
	"vs",
	"xcode",
}

func IsChromaTheme(theme string) bool {
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

func (conf *ServiceConfig) SetTheme(theme string) {
	if theme != LIGHT_THEME && theme != DARK_THEME {
		return
	}

	conf.Theme = theme
}

func (conf *ServiceConfig) SetCodeBlockTheme(codeBlockStyle string) error {
	// User didn't supply option (default is "")
	// This will default to `ServiceConfig` CodeBlockTheme default value.
	if len(codeBlockStyle) == 0 {
		return nil
	}

	if !IsChromaTheme(codeBlockStyle) {
		message := "Unknown theme. The following styles are available:\n\n"
		for _, th := range themes {
			message += "	" + th + "\n"
		}

		return errors.New(message)
	}

	conf.CodeBlockTheme = codeBlockStyle
	return nil
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
		return nil, errors.New(invalid_config_error + err.Error())
	}

	if conf.Theme == "" {
		conf.Theme = DEFAULT
	}
	if conf.CodeBlockTheme == "" || !IsChromaTheme(conf.CodeBlockTheme) {
		conf.CodeBlockTheme = DEFAULT_CODESTYLE
	}

	return &conf, nil
}
