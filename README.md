# MinTorrent

MinTorrent is a minimalistic, zero dependency torrent client, written in Go (golang).

**BitTorrent protocol**

This BitTorrent client implemented the [original BitTorrent protocol spec](https://www.bittorrent.org/beps/bep_0003.html) from 2001.

**Features**

- Simple, 'no-nonsense' torrent leeching (doesn't support seeding yet)
- Supports `.torrent` files (doesn't support magnet links yet)
- HTTP trackers (no UDP trackers)

Also:
- Single binary
- Cross platform
- Tiny runtime footprint — designed for running on servers and in the future, on resource-constrained environments such as embedded devices.

## Installation

```sh
go get github.com/cedrickchee/min-torrent
```

TODO: provide prebuild binaries (compiled downloads).

## Basic usage

```sh
min-torrent <torrent_file_path.torrent> <output_file_path>

# Example: downloading Arch Linux ISO from
# https://www.archlinux.org/download/
min-torrent archlinux-2020.01.01-x86_64.iso.torrent archlinux.iso
```

## Development

### Running on embedded devices/microcontroller boards

That's the plan! However, this is still work-in-progress.

I believe, if we compile the program using [TinyGo](https://tinygo.org/), we can run on several different microcontroller boards such as the Arduino Uno and the Rasberry Pi.

## Limitations

- No support for DHT, uTP, PEX and various extensions.

---

## Why I started this project?

In 2019, I was motivated to learn a new programming language after I've been writing JavaScript for a over a decade. I decided to learn a statically typed and compiled programming language and I choose Go.

I started this project to relearn Go and to practice systems programming. My project idea was inspired by several blog posts and tutorials below (credits to them):

- [A BitTorrent client in Go – Part 1: Torrent file and announcement](https://halfbyte.io/a-bittorrent-client-in-go-part-1-torrent-file-and-announcement/)
- [How to make your own bittorrent client](https://allenkim67.github.io/programming/2016/05/04/how-to-make-your-own-bittorrent-client.html)
- [Building a BitTorrent client from the ground up in Go](https://blog.jse.li/posts/torrent/)
- [A BitTorrent client in Python 3.5](https://markuseliasson.se/article/bittorrent-in-python/)
