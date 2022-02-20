package config

import (
	"regexp"
	"testing"
)

func TestMatchPathToFileImage(t *testing.T) {
	regex, err := regexp.Compile(ImageRegex)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	cases := []struct {
		pattern  string
		expected bool
	}{
		// should match
		{"/pichu.gif", true},
		{"/pichu-&()?!\\.gif", true},
		{"/assets/pichu.gif", true},
		{"/assets/sss/pichu.gif", true},
		{"/pichu.jpg", true},
		{"/pichu.png", true},

		// should not match
		{"/pichu", false},     // no ext
		{"/pichu.pnx", false}, // wrong ext
		{"pichu.gif", false},  // no preceding slash
		{"/abc/pichu.gi", false},
	}

	for _, c := range cases {
		got := regex.Match([]byte(c.pattern))
		if got != c.expected {
			t.Errorf("for \"%s\"; got %t; want %t", c.pattern, got, c.expected)
			t.FailNow()
		}
	}
}
