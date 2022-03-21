# spamd :fire:

spamd is a tool which allows you to **mass preview** Github flavoured markdowns before
commiting them.

![demo](/assets/demo.png)

## Features

* Preview rendered markdowns as you edit
* Open multiple markdown documents easily (using your default browser)
* Only render contents when you visit tab/window
* Can change code block color theme :rainbow:
* Light/Dark toggle :sunny:/:new_moon:
* Auto-close tabs when the server is closed

## Install

### macOS/Linux

Install the latest version for your system:

```sh
curl -sS https://storage.googleapis.com/spamd-releases/install.sh | sh
```

### Windows

Download the release [package](https://github.com/Vui-Chee/spamd/releases/download/v0.1.1/spamd_windows_amd64)
directly.

### Using Go

If you have already installed [go](https://go.dev/dl/), you can run `go get github.com/vui-chee/spamd` or
`go install github.com/vui-chee/spamd@latest`.

## Usage

Run :point_right: `spamd`. This by default opens `README.md` if it exists in the current working directory.

Otherwise, if you want to open specific markdown file(s), you can do so by:

`spamd [file1.md] [file2.md] ...`

For all other features, run `spamd --help`.

## Development

### Run

`go run spamd.go`

### Building

`go build -ldflags="-s -w"`

### Frontend 

The static frontend files used will be *embedded* inside `service/frontend`. The css
is generated with [generate-github-markdown-css](https://github.com/sindresorhus/generate-github-markdown-css) package along with customizations.

## Contributing

1. Check the open issues or open a new issue to start a discussion around your feature idea or the bug you found
2. Fork the repository and make your changes
3. Open a new pull request
