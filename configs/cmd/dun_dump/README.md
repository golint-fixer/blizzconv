dun_dump
========

dun_dump is a tool for constructing dungeons, based on the information retrieved
from a given DUN file, and storing these dungeons as PNG images.

Installation
------------

	$ go get github.com/mewrnd/blizzconv/configs/cmd/dun_dump

Usage
-----

	$ mkdir blizzdump/
	$ cd blizzdump/
	$ ln -s /path/to/extracted/diabdat_mpq/ mpqdump
	$ ln -s $GOPATH/src/github.com/mewrnd/blizzconv/mpq/mpq.ini
	$ ln -s $GOPATH/src/github.com/mewrnd/blizzconv/images/imgconf/cel.ini
	$ ln -s $GOPATH/src/github.com/mewrnd/blizzconv/configs/dunconf/dun.ini
	$ dun_dump -a
