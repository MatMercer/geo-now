package main

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"math"
	"os"
	"os/exec"
	"slices"
	"strings"
	"sync"
	"time"
)

func main() {
	src := "HS_H09_20231130_0030_B03_FLDK_R05"
	dir := "./sample-data"
	sections, err := openFiles(dir, src)
	downsample := 1
	if err != nil {
		fmt.Printf("Failed to open himawari sections: %s\n", err)
		return
	}
	img, err := himawariDecode(sections, downsample)
	if err != nil {
		fmt.Printf("Failed to decode file: %s\n", err)
		return
	}

	fileName := src + fmt.Sprintf("_T%d", time.Now().Unix()) + ".jpg"
	fimg, _ := os.Create(fileName)
	fmt.Printf("Saving to %s...\n", fileName)
	err = jpeg.Encode(fimg, img, &jpeg.Options{Quality: 90})
	if err != nil {
		panic(err)
	}
	if err = fimg.Close(); err != nil {
		panic(err)
	}

	_ = exec.Command("explorer.exe", fileName).Run()
}

// openFiles Returns a list of file sections sorted asc
func openFiles(dir string, pattern string) ([]io.ReadSeekCloser, error) {
	var filesWithPattern []string
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("error reading %q directory: %s", dir, err)
	}
	for _, file := range files {
		if strings.HasPrefix(file.Name(), pattern) {
			filesWithPattern = append(filesWithPattern, dir+"/"+file.Name())
		}
	}
	slices.Sort(filesWithPattern)
	var oFiles []io.ReadSeekCloser
	for _, f := range filesWithPattern {
		ff, err := os.Open(f)
		if err != nil {
			return nil, fmt.Errorf("failed to open %q file: %s", f, err)
		}
		oFiles = append(oFiles, ff)
	}

	return oFiles, nil
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

func himawariDecode(sections []io.ReadSeekCloser, downsample int) (*image.RGBA, error) {
	defer func() {
		for _, s := range sections {
			_ = s.Close()
		}
	}()
	var img *image.RGBA

	// Decode first section to gather file info
	firstSection, err := DecodeFile(sections[0])
	if err != nil {
		return nil, fmt.Errorf("failed to decode first section: %s", err)
	}
	totalSections := len(sections)
	d := calculateScaling(firstSection, downsample)
	img = image.NewRGBA(image.Rect(0, 0, d.scaledWidth, d.scaledHeight*totalSections))
	// Continue to other sections
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		decodeSection(firstSection, downsample, d, img)
	}()
	for section := 1; section < totalSections; section++ {
		wg.Add(1)
		// Decode data
		go func(f io.ReadSeeker) {
			defer wg.Done()
			h, err := DecodeFile(f)
			err = decodeSection(h, downsample, d, img)
			// TODO: err check
			if err != nil {
				//return nil, err
			}
		}(sections[section])
	}
	wg.Wait()

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
