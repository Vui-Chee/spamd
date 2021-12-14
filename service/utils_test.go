package service

import (
	"reflect"
	"testing"
)

func TestFormatStreamData(t *testing.T) {
	inputs := []string{
		"",
		"\n",
		"abc",
		"abc\n",
		"abc\ndef",
		"abc\ndef\n",
	}

	expected := []string{
		"",
		"",
		"data: abc\n\n",
		"data: abc\n\n",
		"data: abc\ndata: def\n\n",
		"data: abc\ndata: def\n\n",
	}

	for i, input := range inputs {
		got := eventStreamFormat(input)
		if !reflect.DeepEqual(got, []byte(expected[i])) {
			t.Errorf("case %d, eventStreamFormat returns \"%s\", expected \"%s\"", i+1, got, expected[i])
		}
	}
}
