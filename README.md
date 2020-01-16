# LeechTorrent

LeechTorrent is a simple CLI for downloading torrent. Its tiny footprint was designed for running on resource-constrained environments such as embedded devices and servers. The torrent client was implemented in Go programming language.

**BitTorrent protocol**

This BitTorrent client implemented the [original BitTorrent protocol spec](https://www.bittorrent.org/beps/bep_0003.html) from 2001.

**Features**

- Supports `.torrent` files (no magnet links)
- Supports HTTP trackers (no UDP trackers)
- Leeches torrent (does not support uploading pieces)

## Installation

```sh
go get github.com/cedrickchee/leechtorrent
```

TODO: provide prebuild binaries (compiled downloads).

## Basic usage

```sh
leechtorrent <torrent_file_path.torrent> <output_file_path>

# Example: downloading Arch Linux ISO from
# https://www.archlinux.org/download/
leechtorrent archlinux-2020.01.01-x86_64.iso.torrent archlinux.iso
```

## Development

### Running on embedded devices/microcontroller boards

That's the plan! However, this is still work-in-progress.

I believe, if we compile the program using [TinyGo](https://tinygo.org/), we can run on several different microcontroller boards such as the Arduino Uno and the Rasberry Pi.

---

## Why I started this project?

In 2019, I was motivated to learn a new programming language after I've been writing JavaScript for a over a decade. I decided to learn a statically typed and compiled programming language and I choose Go.

I started this project to relearn Go and to practice systems programming. My project idea was inspired by several blog posts and tutorials below (credits to them):

- [A BitTorrent client in Go – Part 1: Torrent file and announcement](https://halfbyte.io/a-bittorrent-client-in-go-part-1-torrent-file-and-announcement/)
- [How to make your own bittorrent client](https://allenkim67.github.io/programming/2016/05/04/how-to-make-your-own-bittorrent-client.html)
- [Building a BitTorrent client from the ground up in Go](https://blog.jse.li/posts/torrent/)
- [A BitTorrent client in Python 3.5](https://markuseliasson.se/article/bittorrent-in-python/)
