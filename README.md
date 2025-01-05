[![Go Reference](https://pkg.go.dev/badge/github.com/sauromates/leech.svg)](https://pkg.go.dev/github.com/sauromates/leech)

# LEECH

Console BitTorrent client written in Go.

## Usage

1. Clone repository
2. Compile locally or run with `go run main.go <torrent_file>`

Downloaded files will be saved into a subdirectory named after torrent itself.
Support of custom download directories is on the roadmap.

## Features

Leech currently supports only the most simple download via `.torrent` files.

Uploading, magnet links and DHT are not supported for now.

## Acknowledgements

Thanks to amazing [article](https://blog.jse.li/posts/torrent/) by the author
of [this repository](https://github.com/veggiedefender/torrent-client).
