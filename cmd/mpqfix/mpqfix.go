package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
)

var mpqpath string

func init() {
	flag.StringVar(&mpqpath, "mpqdump", "mpqdump/", "Path to an extracted MPQ file.")
}

func main() {
	flag.Parse()
	fixes := []Fix{
		{
			path:   "levels/l1data/banner2.dun",
			data:   map[int]byte{6: 0x02, 8: 0x02, 12: 0x02, 14: 0x02, 16: 0x02},
			oldsum: [md5.Size]byte{0x10, 0x1C, 0x30, 0x9E, 0xCC, 0x06, 0xE2, 0x49, 0xE0, 0xFF, 0x14, 0xD9, 0x9E, 0x4D, 0xB8, 0x61},
			newsum: [md5.Size]byte{0xA5, 0x19, 0xF7, 0x38, 0xA5, 0xCC, 0xE3, 0xD5, 0x04, 0x41, 0xD6, 0x21, 0xDE, 0x35, 0x01, 0x69},
		},
		{
			path:   "monsters/darkmage/dmagew.cl2",
			data:   map[int]byte{4: 0x28, 8: 0x30, 12: 0x38, 16: 0x40, 20: 0x48, 24: 0x50, 28: 0x58},
			oldsum: [md5.Size]byte{0xB8, 0x98, 0x12, 0x7A, 0x89, 0x98, 0x44, 0x90, 0xB7, 0x2C, 0xA9, 0x92, 0x56, 0xF1, 0x44, 0x15},
			newsum: [md5.Size]byte{0x2C, 0x8B, 0xE0, 0x23, 0xB8, 0x28, 0xA2, 0x17, 0xA4, 0x7E, 0x15, 0x9F, 0x30, 0xC0, 0x44, 0x33},
		},
		{
			path:   "monsters/unrav/unravw.cel",
			data:   map[int]byte{4: 0x07, 8: 0xC5, 12: 0xA3, 16: 0xC3, 20: 0x26, 24: 0x4C, 28: 0x93},
			oldsum: [md5.Size]byte{0x09, 0xAE, 0xC6, 0x35, 0xE0, 0xFC, 0x9E, 0x08, 0x43, 0x91, 0xF0, 0x0D, 0x4C, 0xDD, 0xB2, 0x99},
			newsum: [md5.Size]byte{0x59, 0xE5, 0x2E, 0x32, 0xA7, 0x35, 0xCC, 0x46, 0x42, 0x6A, 0x36, 0xB2, 0x40, 0x7B, 0xE4, 0xC0},
		},
	}
	for _, fix := range fixes {
		fmt.Printf("Patching %q.\n", fix.path)
		err := fix.Apply()
		if err != nil {
			log.Println(err)
			continue
		}
	}
}

type Fix struct {
	path   string
	data   map[int]byte
	oldsum [md5.Size]byte
	newsum [md5.Size]byte
}

func (fix Fix) Apply() error {
	path := filepath.Join(mpqpath, fix.path)
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	oldsum := md5.Sum(buf)
	if oldsum == fix.newsum {
		return fmt.Errorf("%q already patched.", fix.path)
	}
	if oldsum != fix.oldsum {
		return fmt.Errorf("MD5 checksum mismatch for unpatched version of %q.", fix.path)
	}
	err = ioutil.WriteFile(path+".orig", buf, 0644)
	if err != nil {
		return fmt.Errorf("failed to create backup for %q; %v", fix.path, err)
	}
	for pos, val := range fix.data {
		buf[pos] = val
	}
	newsum := md5.Sum(buf)
	if newsum != fix.newsum {
		return fmt.Errorf("MD5 checksum mismatch for patched version of %q.", fix.path)
	}
	return ioutil.WriteFile(path, buf, 0644)
}
