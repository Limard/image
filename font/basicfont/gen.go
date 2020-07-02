// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

// This program generates data.go.
package main

import (
	"bytes"
	"fmt"
	"go/format"
	"image"
	"image/draw"
	"io/ioutil"
	"log"
	"path"
	"path/filepath"

	"github.com/Limard/image/font"
	"github.com/Limard/image/font/plan9font"
	"github.com/Limard/image/math/fixed"
)

func main() {
	// nGlyphs is the number of glyphs to generate: 95 characters in the range
	// [0x20, 0x7e], plus the replacement character.
	const nGlyphs = 95 + 1
	// The particular font (unicode.7x13.font) leaves the right-most column
	// empty in its ASCII glyphs. We don't have to include that column in the
	// generated glyphs, so we subtract one off the effective width.
	const width, height, ascent = 7 - 1, 13, 11

	readFile := func(name string) ([]byte, error) {
		return ioutil.ReadFile(filepath.FromSlash(path.Join("../testdata/fixed", name)))
	}
	fontData, err := readFile("unicode.7x13.font")
	if err != nil {
		log.Fatalf("readFile: %v", err)
	}
	face, err := plan9font.ParseFont(fontData, readFile)
	if err != nil {
		log.Fatalf("plan9font.ParseFont: %v", err)
	}

	dst := image.NewRGBA(image.Rect(0, 0, width, nGlyphs*height))
	draw.Draw(dst, dst.Bounds(), image.Black, image.Point{}, draw.Src)
	d := &font.Drawer{
		Dst:  dst,
		Src:  image.White,
		Face: face,
	}
	for i := 0; i < nGlyphs; i++ {
		r := '\ufffd'
		if i < nGlyphs-1 {
			r = 0x20 + rune(i)
		}
		d.Dot = fixed.P(0, height*i+ascent)
		d.DrawString(string(r))
	}

	w := bytes.NewBuffer(nil)
	w.WriteString(preamble)
	fmt.Fprintf(w, "// mask7x13 contains %d %d×%d glyphs in %d Pix bytes.\n", nGlyphs, width, height, nGlyphs*width*height)
	fmt.Fprintf(w, "var mask7x13 = &image.Alpha{\n")
	fmt.Fprintf(w, "  Stride: %d,\n", width)
	fmt.Fprintf(w, "  Rect: image.Rectangle{Max: image.Point{%d, %d*%d}},\n", width, nGlyphs, height)
	fmt.Fprintf(w, "  Pix: []byte{\n")
	b := dst.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		if y%height == 0 {
			if y != 0 {
				w.WriteByte('\n')
			}
			i := y / height
			if i < nGlyphs-1 {
				i += 0x20
				fmt.Fprintf(w, "// %#2x %q\n", i, rune(i))
			} else {
				fmt.Fprintf(w, "// U+FFFD REPLACEMENT CHARACTER\n")
			}
		}

		for x := b.Min.X; x < b.Max.X; x++ {
			if dst.RGBAAt(x, y).R > 0 {
				w.WriteString("0xff,")
			} else {
				w.WriteString("0x00,")
			}
		}
		w.WriteByte('\n')
	}
	w.WriteString("},\n}\n")

	fmted, err := format.Source(w.Bytes())
	if err != nil {
		log.Fatalf("format.Source: %v", err)
	}
	if err := ioutil.WriteFile("data.go", fmted, 0644); err != nil {
		log.Fatalf("ioutil.WriteFile: %v", err)
	}
}

const preamble = `// generated by go generate; DO NOT EDIT.

package basicfont

// This data is derived from files in the font/fixed directory of the Plan 9
// Port source code (https://github.com/9fans/plan9port) which were originally
// based on the public domain X11 misc-fixed font files.

import "image"

`
