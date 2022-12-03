package options

import (
	"flag"

	"spamd/internal/sys"
)

const (
	beginUsage = "Usage: spamd [options...] <path-to-markdown>\nOptions:"
	endUsage   = `Additionally, if you want to persist any of this configs, you can
create a .spamd JSON file at your ROOT directory containing:

	{
	  "theme": "dark",
	  "codeblock": "fruity",
	  "port": 3000
	}

This is just an example. You can change/omit any of the fields.
`
)

type Options struct {
	ShowVersion bool
	NoBrowser   bool
	Port        int
	Theme       string
	CodeStyle   string
}

func ParseOptions() *Options {
	options := &Options{}
	flag.BoolVar(&options.ShowVersion, "v", false, "Display version and exit")
	flag.BoolVar(&options.NoBrowser, "nb", false, "Do not open browser if this is set true (default: false)")
	flag.IntVar(&options.Port, "p", 0, "Port number (fixed port, otherwise a RANDOM port is supplied)")
	flag.StringVar(&options.Theme, "t", "", "Display markdown HTML in \"dark\" or \"light\" theme. (default: light)")
	flag.StringVar(&options.CodeStyle, "c", "", "The style you want to apply to your code blocks. (default: monokai)")
	flag.Usage = func() {
		sys.Eprintf("%s\n\n", beginUsage)
		flag.PrintDefaults()
		sys.Eprintf("\n%s", endUsage)
	}
	flag.Parse()
	return options
}
