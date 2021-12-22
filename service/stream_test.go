package service

import (
	"reflect"
	"testing"
)

func TestFormatStreamData(t *testing.T) {
	// Each data packet must end with newline.
	// A single newline is also a data packet - empty data packet.
	inputs := []string{
		"",
		"\n",
		"abc",
		"abc\n",
		"abc\ndef",
		"abc\ndef\n",

		// consecutive newlines
		"abc\n\n",
		"abc\n\ndef",
		"abc\n\n\ndef",
		"abc\n\ndef\n",
	}

	expected := []string{
		"",
		"data: \ndata: \n\n",
		"data: abc\n\n",
		"data: abc\ndata: \n\n",
		"data: abc\ndata: def\n\n",
		"data: abc\ndata: def\ndata: \n\n",

		"data: abc\ndata: \ndata: \n\n",
		"data: abc\ndata: \ndata: def\n\n",
		"data: abc\ndata: \ndata: \ndata: def\n\n",
		"data: abc\ndata: \ndata: def\ndata: \n\n",
	}

	for i, input := range inputs {
		got := eventStreamFormat(input)
		if !reflect.DeepEqual(got, []byte(expected[i])) {
			t.Errorf("case %d, eventStreamFormat returns \"%s\", expected \"%s\"", i+1, got, expected[i])
		}
	}
}
