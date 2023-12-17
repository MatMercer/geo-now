package main

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"math"
	"os"
	"os/exec"
)

func main() {
	err := himawariDecode(5)

	if err != nil {
		fmt.Printf("Failed to decode file: %s\n", err)
		return
	}
	_ = exec.Command("explorer.exe", "image.jpg").Run()
}

func himawariDecode(downsample int) interface{} {
	// TODO: extract into multiple functions
	// opening the files through a pattern
	// assuming 10 as the section count
	// decoding the header of each section
	// joining each section to an *image scaling if necessary
	// encoding to jpeg
	src := "HS_H09_20231130_0030_B04_FLDK_R10"
	filePattern := "sample-data/%s_S%02d10.DAT"
	sectionCount := 10
	var img *image.RGBA
	for section := 0; section < sectionCount; section++ {
		// Decode data
		f, err := os.Open(fmt.Sprintf(filePattern, src, section+1))
		if err != nil {
			panic(err)
		}
		h, err := DecodeFile(f)
		width := int(h.DataInfo.NumberOfColumns)
		height := int(h.DataInfo.NumberOfLines)
		scale := 1.0 / float64(downsample)
		scaledWidth := int(scale * float64(width))
		scaledHeight := int(scale * float64(height))
		// Start and End Y are the relative positions for the final image based in a section
		startY := scaledHeight * section
		endY := startY + scaledHeight
		// Amount of pixels for down sample skip
		skipPx := downsample - 1
		// Initialize image if first loop
		if section == 0 {
			img = image.NewRGBA(image.Rect(0, 0, scaledWidth, scaledHeight*sectionCount))
		}
		fmt.Printf("Decoding %dx%d from y %d-%d\n", width, height, startY, endY)
		for y := startY; y < endY; y++ {
			for x := 0; x < scaledWidth; x++ {
				// Do err and outside scan area logic
				err := readPixel(h, img, x, y)
				if err != nil {
					return err
				}
				err = h.Seek(skipPx)
				if err != nil {
					return fmt.Errorf("failed to skip %d pixels at %d:%d: %skipPx", skipPx, x, y, err)
				}
			}
			err = h.Seek(width * skipPx)
			if err != nil {
				return fmt.Errorf("failed to skip %d pixels at %d:%d: %skipPx", skipPx, 0, y, err)
			}
		}
		_ = f.Close()
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

func readPixel(h *HMFile, img *image.RGBA, x int, y int) error {
	data, err := h.ReadPixel()
	if err != nil {
		return fmt.Errorf("failed to read pixel at %d:%d: %s", x, y, err)
	}
	if data == h.CalibrationInfo.CountValueOfPixelsOutsideScanArea || data == h.CalibrationInfo.CountValueOfErrorPixels {
		img.Set(x, y, color.Black)
		return nil
	}

	// Get a number between 0 and 1 from max number of pixels
	// different bands has different number of pixels, e.g., band 03 has 11
	coef := float64(data) / (math.Pow(2., float64(h.CalibrationInfo.ValidNumberOfBitsPerPixel)) - 2.)
	pc := pixel(coef, 1)
	img.Set(x, y, color.RGBA{R: uint8(pc), G: uint8(pc), B: uint8(pc), A: 255})

	return nil
}

// pixel Returns 255*coef clamping at coef, brightness adjusted
func pixel(coef, brig float64) int {
	return int(math.Min(coef*255*brig, 255))
}
