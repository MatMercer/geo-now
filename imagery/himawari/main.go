package main

import (
	"compress/bzip2"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"math"
	"os"
	"os/exec"
)

func main() {
	err := himawariDecode(8)

	if err != nil {
		fmt.Printf("Failed to decode file: %s\n", err)
		return
	}
	_ = exec.Command("explorer.exe", "image.jpg").Run()
}

func himawariDecode(downsample int) interface{} {
	src := "HS_H09_20231130_0030_B04_FLDK_R10"
	filePattern := "sample-data/%s_S%02d10.DAT.bz2"
	sectionCount := 10
	var img *image.RGBA
	for section := 0; section < sectionCount; section++ {
		// Decode data
		f, err := os.Open(fmt.Sprintf(filePattern, src, section+1))
		if err != nil {
			panic(err)
		}
		h, err := DecodeFile(bzip2.NewReader(f))
		//_ = f.Close()
		if err != nil {
			panic(err)
		}
		width := int(h.DataInfo.NumberOfColumns)
		height := int(h.DataInfo.NumberOfLines)
		scale := 1.0 / float64(downsample)
		scaledWidth := int(scale * float64(width))
		scaledHeight := int(scale * float64(height))
		// Start and End Y are the relative positions for the final image based in a section
		startY := scaledHeight * section
		endY := startY + scaledHeight
		// Initialize image if first loop
		if section == 0 {
			img = image.NewRGBA(image.Rect(0, 0, scaledWidth, scaledHeight*sectionCount))
		}
		fmt.Printf("Decoding %dx%d from y %d-%d\n", width, height, startY, endY)
		d := 0
		for y, l := startY, 1.0; y < endY; y, l = y+1, l+1 {
			for x := 0; x < scaledWidth; x++ {
				// Do err and outside scan area logic
				rawData, err := h.ReadPixel()
				if err != nil {
					return fmt.Errorf("failed to read pixel at %d:%d: %s", x, y, err)
				}
				pToImage(h, rawData, img, x, y)
				h.Seek(downsample - 1)
				d += downsample - 1
			}
			h.Seek(width * (downsample - 1))
		}
	}
	fimg, _ := os.Create("image.jpg")
	err := jpeg.Encode(fimg, img, &jpeg.Options{Quality: 90})
	if err != nil {
		panic(err)
	}
	if err = fimg.Close(); err != nil {
		return err
	}

	return nil
}

func pToImage(h *HMFile, data uint16, img *image.RGBA, x int, y int) {
	// Err check
	if data == h.CalibrationInfo.CountValueOfPixelsOutsideScanArea || data == h.CalibrationInfo.CountValueOfErrorPixels {
		img.Set(x, y, color.Black)
		return
	}
	// Get a number between 0 and 1 from max number of pixels
	// different bands has different number of pixels, e.g., band 03 has 11
	coef := float64(data) / (math.Pow(2., float64(h.CalibrationInfo.ValidNumberOfBitsPerPixel)) - 2.)
	pc := pixel(coef, 1)
	img.Set(x, y, color.RGBA{R: uint8(pc), G: uint8(pc), B: uint8(pc), A: 255})
}

// pixel Returns 255*coef clamping at coef, brightness adjusted
func pixel(coef, brig float64) int {
	return int(math.Min(coef*255*brig, 255))
}
