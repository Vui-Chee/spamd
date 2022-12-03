package browser

import (
	"flag"
	"os/exec"

	"spamd/internal/sys"
)

const (
	defaultMarkdown = "README.md"
)

type BrowserDelegate struct {
	cmd *exec.Cmd
}

func (b BrowserDelegate) Start() error {
	return b.cmd.Start()
}

func Commands(url string) sys.Commands {
	return sys.Commands{
		"linux":   BrowserDelegate{cmd: exec.Command("xdg-open", url)},
		"windows": BrowserDelegate{cmd: exec.Command("rundll32", "url.dll,FileProtocolHandler", url)},
		"darwin":  BrowserDelegate{cmd: exec.Command("open", url)},
	}
}

func MassOpen(baseUrl string, nobrowser bool) {
	var filepath string = defaultMarkdown
	if flag.NArg() >= 1 {
		for i := 0; i < len(flag.Args()); i++ {
			filepath := flag.Args()[i]

			if !sys.IsFileWithExt(filepath, ".md") {
				sys.Eprintf("%s is not a markdown document.\n", filepath)
			} else if !sys.Exists(filepath) {
				sys.Eprintf("%s does not exist.\n", filepath)
			} else {
				if !nobrowser {
					go func() {
						sys.Exec(Commands(baseUrl + "/" + filepath))
					}()
				}
			}
		}
	} else {
		if !nobrowser && sys.IsFileWithExt(filepath, ".md") && sys.Exists(filepath) {
			sys.Exec(Commands(baseUrl + "/" + filepath))
		}
	}
}
