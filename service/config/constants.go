package config

import "fmt"

const (
	RefreshPrefix = "/__/refresh"
	StylesPrefix  = "/__/styles"
)

func RefreshPattern() string {
	return fmt.Sprintf("^%s/.+", RefreshPrefix)
}
