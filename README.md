# LeechTorrent

LeechTorrent is a CLI for downloading torrent. Its tiny footprint was designed for running on resource-constrained environments such as embedded devices and servers. The torrent client was implemented in Go programming language.

Back in Dec 2019, I started this project to learn Go and systems programming.

**BitTorrent protocol**

This BitTorrent client implemented the [original BitTorrent protocol spec](https://www.bittorrent.org/beps/bep_0003.html) from 2001.

**Features**

- Supports `.torrent` files (no magnet links)
- Supports HTTP trackers (no UDP trackers)
- Leeches torrent (does not support uploading pieces)

## Download

```sh
go get github.com/cedrickchee/leechtorrent
```

_prebuild binaries are coming soon..._

## How to use

```sh
leechtorrent <torrent_file_path.torrent> <output_file_path>

# Example: downloading Arch Linux ISO from
# https://www.archlinux.org/download/
leechtorrent archlinux-2020.01.01-x86_64.iso.torrent archlinux.iso
```

## Development

### Running on embedded devices/microcontroller boards

 Compile the program using [TinyGo](https://tinygo.org/) so you can run on several different microcontroller boards such as the Arduino Uno and the Rasberry Pi.

### Test Data

Download test data from ... and copy into the `./torrentfile/test_data` directory before you run your test suite using `go test ./...`.
