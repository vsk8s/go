// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ignore
// +build ignore

package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"log"
	"os"
)

var debug = flag.Bool("debug", false, "")

func main() {
	flag.Parse()

	w := new(bytes.Buffer)
	w.WriteString(pre)
	for _, sratio := range subsampleRatios {
		fmt.Fprintf(w, sratioCase, sratio, sratioLines[sratio])
	}
	w.WriteString(post)

	if *debug {
		os.Stdout.Write(w.Bytes())
		return
	}
	out, err := format.Source(w.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile("impl.go", out, 0660); err != nil {
		log.Fatal(err)
	}
}

const pre = `// Code generated by go run gen.go; DO NOT EDIT.

package imageutil

import (
	"image"
)

// DrawYCbCr draws the YCbCr source image on the RGBA destination image with
// r.Min in dst aligned with sp in src. It reports whether the draw was
// successful. If it returns false, no dst pixels were changed.
//
// This function assumes that r is entirely within dst's bounds and the
// translation of r from dst coordinate space to src coordinate space is
// entirely within src's bounds.
func DrawYCbCr(dst *image.RGBA, r image.Rectangle, src *image.YCbCr, sp image.Point) (ok bool) {
	// This function exists in the image/internal/imageutil package because it
	// is needed by both the image/draw and image/jpeg packages, but it doesn't
	// seem right for one of those two to depend on the other.
	//
	// Another option is to have this code be exported in the image package,
	// but we'd need to make sure we're totally happy with the API (for the
	// rest of Go 1 compatibility), and decide if we want to have a more
	// general purpose DrawToRGBA method for other image types. One possibility
	// is:
	//
	// func (src *YCbCr) CopyToRGBA(dst *RGBA, dr, sr Rectangle) (effectiveDr, effectiveSr Rectangle)
	//
	// in the spirit of the built-in copy function for 1-dimensional slices,
	// that also allowed a CopyFromRGBA method if needed.

	x0 := (r.Min.X - dst.Rect.Min.X) * 4
	x1 := (r.Max.X - dst.Rect.Min.X) * 4
	y0 := r.Min.Y - dst.Rect.Min.Y
	y1 := r.Max.Y - dst.Rect.Min.Y
	switch src.SubsampleRatio {
`

const post = `
	default:
		return false
	}
	return true
}
`

const sratioCase = `
	case image.YCbCrSubsampleRatio%s:
		for y, sy := y0, sp.Y; y != y1; y, sy = y+1, sy+1 {
			dpix := dst.Pix[y*dst.Stride:]
			yi := (sy-src.Rect.Min.Y)*src.YStride + (sp.X - src.Rect.Min.X)
			%s

				// This is an inline version of image/color/ycbcr.go's func YCbCrToRGB.
				yy1 := int32(src.Y[yi]) * 0x10101
				cb1 := int32(src.Cb[ci]) - 128
				cr1 := int32(src.Cr[ci]) - 128

				// The bit twiddling below is equivalent to
				//
				// r := (yy1 + 91881*cr1) >> 16
				// if r < 0 {
				//     r = 0
				// } else if r > 0xff {
				//     r = ^int32(0)
				// }
				//
				// but uses fewer branches and is faster.
				// Note that the uint8 type conversion in the return
				// statement will convert ^int32(0) to 0xff.
				// The code below to compute g and b uses a similar pattern.
				r := yy1 + 91881*cr1
				if uint32(r)&0xff000000 == 0 {
					r >>= 16
				} else {
					r = ^(r >> 31)
				}

				g := yy1 - 22554*cb1 - 46802*cr1
				if uint32(g)&0xff000000 == 0 {
					g >>= 16
				} else {
					g = ^(g >> 31)
				}

				b := yy1 + 116130*cb1
				if uint32(b)&0xff000000 == 0 {
					b >>= 16
				} else {
					b = ^(b >> 31)
				}


				// use a temp slice to hint to the compiler that a single bounds check suffices
				rgba := dpix[x : x+4 : len(dpix)]
				rgba[0] = uint8(r)
				rgba[1] = uint8(g)
				rgba[2] = uint8(b)
				rgba[3] = 255
			}
		}
`

var subsampleRatios = []string{
	"444",
	"422",
	"420",
	"440",
}

var sratioLines = map[string]string{
	"444": `
		ci := (sy-src.Rect.Min.Y)*src.CStride + (sp.X - src.Rect.Min.X)
		for x := x0; x != x1; x, yi, ci = x+4, yi+1, ci+1 {
	`,
	"422": `
		ciBase := (sy-src.Rect.Min.Y)*src.CStride - src.Rect.Min.X/2
		for x, sx := x0, sp.X; x != x1; x, sx, yi = x+4, sx+1, yi+1 {
			ci := ciBase + sx/2
	`,
	"420": `
		ciBase := (sy/2-src.Rect.Min.Y/2)*src.CStride - src.Rect.Min.X/2
		for x, sx := x0, sp.X; x != x1; x, sx, yi = x+4, sx+1, yi+1 {
			ci := ciBase + sx/2
	`,
	"440": `
		ci := (sy/2-src.Rect.Min.Y/2)*src.CStride + (sp.X - src.Rect.Min.X)
		for x := x0; x != x1; x, yi, ci = x+4, yi+1, ci+1 {
	`,
}
