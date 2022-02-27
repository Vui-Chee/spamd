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

## Download

If you have already installed [go](https://go.dev/dl/), you can run `go get github.com/vui-chee/spamd` or
`go install github.com/vui-chee/spamd@latest`.

## Usage

Run :point_right: `spamd`. This by default opens `README.md` if it exists in the current working directory.

Otherwise, if you want to open specific markdown file(s), you can do so by:

`spamd [file1.md] [file2.md] ...`

For all other features, run `spamd --help`.

```
Usage: spamd [options...] <path-to-markdown>

Options:
        -p Port number (fixed port, otherwise a RANDOM port is supplied)
        -t Display markdown HTML in "dark" or "light" theme. (default: light)
        -c The style you want to apply to your code blocks. (default: monokai)
        -nb Do not open browser if this is set true (default: false)

Additionally, if you want to persist any of this configs, you can
create a .spamd JSON file at your ROOT directory containing:

        {
          "theme": "dark",
          "codeblock": "fruity",
          "port": 3000
        }

This is just an example. You can change/omit any of the fields.
```

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
