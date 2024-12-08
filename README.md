# LEECH

Console BitTorrent client written in Go.

## Usage

1. Clone repository
2. Compile locally or run with `go run main.go <torrent_file> <target_file>`

## Features

Leech currently supports only the most simple download via `.torrent` files. It reads the whole thing
into memory so please don't try to download anything that exceeds your memory capacity.

Uploading, magnet links and DHT are not supported for now.

## Acknowledgements

Thanks to amazing [article](https://blog.jse.li/posts/torrent/) by the author of [this repository](https://github.com/veggiedefender/torrent-client).
