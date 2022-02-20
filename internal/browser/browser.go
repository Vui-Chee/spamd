package browser

import (
	"os/exec"

	"github.com/vui-chee/spamd/internal/sys"
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
