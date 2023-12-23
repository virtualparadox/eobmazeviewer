package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	"image/color"
)

func ConvertPalettedToRGBA(palettedImage *image.Paletted, transparent bool) *image.RGBA {
	// Create a new RGBA image with the same dimensions
	rgbaImage := image.NewRGBA(palettedImage.Rect)

	for y := 0; y < palettedImage.Bounds().Dy(); y++ {
		for x := 0; x < palettedImage.Bounds().Dx(); x++ {
			// Set pixel to some color with alpha value
			index := palettedImage.Pix[palettedImage.PixOffset(x, y)]
			indexedColor := palettedImage.Palette[index]
			r, g, b, _ := indexedColor.RGBA()
			a := 254
			if transparent && r == 0 && g == 0 && b == 0 {
				a = 0
			}
			rgbaImage.SetRGBA(x, y, color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a)}) // A = 254 to force 32 bit
		}
	}

	return rgbaImage
}

func BytesToPalettedImage(data *[]byte, width, height int, palette color.Palette) *image.Paletted {
	img := image.NewPaletted(image.Rect(0, 0, width, height), palette)
	offset := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			d := *data
			img.SetColorIndex(x, y, d[offset])
			offset++
		}
	}
	return img
}

func ConvertRGBAtoUint32Array(rgba *image.RGBA) *[]uint32 {
	width, height := rgba.Rect.Dx(), rgba.Rect.Dy()
	pixels := make([]uint32, width*height)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			i := y*rgba.Stride + x*4
			r, g, b, a := rgba.Pix[i], rgba.Pix[i+1], rgba.Pix[i+2], rgba.Pix[i+3]
			pixels[y*width+x] = uint32(r)<<24 | uint32(g)<<16 | uint32(b)<<8 | uint32(a)
		}
	}

	return &pixels
}

func ConvertUint32ArrayToEbitenImage(pixels *[]uint32, width, height int) *ebiten.Image {
	img := ebiten.NewImage(width, height)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			index := y*width + x
			pixel := (*pixels)[index]

			r := uint8((pixel >> 24) & 0xFF)
			g := uint8((pixel >> 16) & 0xFF)
			b := uint8((pixel >> 8) & 0xFF)
			a := uint8(pixel & 0xFF)

			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: a})
		}
	}

	return img
}
