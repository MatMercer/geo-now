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
	img, err := himawariDecode(8)

	if err != nil {
		fmt.Printf("Failed to decode file: %s\n", err)
		return
	}

	fimg, _ := os.Create("image.jpg")
	err = jpeg.Encode(fimg, img, &jpeg.Options{Quality: 90})
	if err != nil {
		panic(err)
	}
	if err = fimg.Close(); err != nil {
		panic(err)
	}

	_ = exec.Command("explorer.exe", "image.jpg").Run()
}

func decodeSection(h *HMFile, downsample int, d sectionDecode, img *image.RGBA) error {
	// Start and End Y are the relative positions for the final image based in a section
	startY := d.scaledHeight * int(h.SegmentInfo.SegmentSequenceNumber-1)
	endY := startY + d.scaledHeight
	// Amount of pixels for down sample skip
	skipPx := downsample - 1
	fmt.Printf("Decoding %dx%d from y %d-%d\n", d.width, d.height, startY, endY)
	for y := startY; y < endY; y++ {
		for x := 0; x < d.scaledWidth; x++ {
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
		err := h.Seek(d.width * skipPx)
		if err != nil {
			return fmt.Errorf("failed to skip %d pixels at %d:%d: %skipPx", skipPx, 0, y, err)
		}
	}
	return nil
}

// Aux struct to store decode metadata
type sectionDecode struct {
	width        int
	height       int
	downsample   int
	scale        float64
	scaledWidth  int
	scaledHeight int
}

func himawariDecode(downsample int) (*image.RGBA, error) {
	// TODO: extract into multiple functions
	// opening the files through a pattern
	src := "HS_H09_20231130_0030_B04_FLDK_R10"
	filePattern := "sample-data/%s_S%02d10.DAT"
	var img *image.RGBA

	// Decode first section to gather file info
	f, err := os.Open(fmt.Sprintf(filePattern, src, 1))
	if err != nil {
		return nil, err
	}
	firstSection, err := DecodeFile(f)
	totalSections := int(firstSection.SegmentInfo.SegmentTotalNumber)
	d := calculateScaling(firstSection, downsample)
	img = image.NewRGBA(image.Rect(0, 0, d.scaledWidth, d.scaledHeight*totalSections))
	err = decodeSection(firstSection, downsample, d, img)
	// Continue to other sections
	for section := 1; section < totalSections; section++ {
		// Decode data
		f, err := os.Open(fmt.Sprintf(filePattern, src, section+1))
		if err != nil {
			panic(err)
		}
		h, err := DecodeFile(f)
		err = decodeSection(h, downsample, d, img)
		if err != nil {
			return nil, err
		}
	}

	return img, nil
}

func calculateScaling(h *HMFile, downsample int) sectionDecode {
	d := sectionDecode{
		width:      int(h.DataInfo.NumberOfColumns),
		height:     int(h.DataInfo.NumberOfLines),
		downsample: downsample,
		scale:      1.0 / float64(downsample),
	}
	d.scaledWidth = int(d.scale * float64(d.width))
	d.scaledHeight = int(d.scale * float64(d.height))

	return d
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
