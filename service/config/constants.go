package config

import "fmt"

const (
	RefreshPrefix = "/__/refresh"
	StylesPrefix  = "/__/styles"

	// Matches these image types.
	ImageRegex = "^\\/.+.(png|jpg|gif|jpeg)$"
)

func RefreshPattern() string {
	return fmt.Sprintf("^%s/.+", RefreshPrefix)
}
