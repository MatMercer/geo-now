package main

import (
	"github.com/djherbis/buffer"
	"github.com/djherbis/nio"
	_ "net/http/pprof"
)

import (
	"fmt"
	"image"
	"image/color"
	"io"
	"math"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"
	"testing"
)

func TestDecode(t *testing.T) {
	// Profiling with pprof
	var wg sync.WaitGroup
	go func() {
		fmt.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	wg.Add(1)

	src := "HS_H09_20231130_0030_B03_FLDK_R05"
	dir := "./sample-data"
	sections, err := openFiles(dir, src)
	if err != nil {
		fmt.Printf("Failed to open himawari sections: %s\n", err)
		return
	}
	_, err = himawariDecode(sections)
	if err != nil {
		fmt.Printf("Failed to decode file: %s\n", err)
		return
	}

	//fileName := src + fmt.Sprintf("_T%d", time.Now().Unix()) + ".jpg"
	//fimg, _ := os.Create(fileName)
	//fmt.Printf("Saving to %s...\n", fileName)
	//err = jpeg.Encode(fimg, img, &jpeg.Options{Quality: 90})
	//if err != nil {
	//	panic(err)
	//}
	//if err = fimg.Close(); err != nil {
	//	panic(err)
	//}
	//
	//_ = exec.Command("explorer.exe", fileName).Run()
	wg.Wait()
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
	width  int
	height int
}

func decodeToFile(h *HMFile, d sectionDecode) error {
	b := buffer.New(32 * 1024)
	r, w := nio.Pipe(b)
	defer w.Close()

	// Open the file for writing
	file, err := os.Create(fmt.Sprintf("section_%d.dat", h.SegmentInfo.SegmentSequenceNumber))
	if err != nil {
		return err
	}
	defer file.Close()

	// Start a goroutine to write to the file
	go func() {
		_, _ = io.Copy(file, r)
	}()

	// Start and End Y are the relative positions for the final image based in a section
	section := h.SegmentInfo.SegmentSequenceNumber
	startY := d.height * int(section-1)
	endY := startY + d.height
	// Amount of pixels for down sample skip
	//skipPx := downsample - 1
	fmt.Printf("Decoding section %d, %dx%d from y %d-%d\n", section, d.width, d.height, startY, endY)
	var pair [2]byte
	var pixelData [1]byte
	for y := startY; y < endY; y++ {
		for x := 0; x < d.width; x++ {
			_, err := h.ImageData.Read(pair[:])
			if err != nil {
				return err
			}
			p := uint16(pair[0]) | uint16(pair[1])<<8
			data := byte(255 * (float64(p) / (math.Pow(2., float64(h.CalibrationInfo.ValidNumberOfBitsPerPixel)) - 2.)))
			pixelData[0] = data
			w.Write(pixelData[:])

			// Do err and outside scan area logic
			//p, err := h.ReadPixel()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func decodeSection(h *HMFile, d sectionDecode, img *image.RGBA) error {
	// Start and End Y are the relative positions for the final image based in a section
	section := h.SegmentInfo.SegmentSequenceNumber
	startY := d.height * int(section-1)
	endY := startY + d.height
	// Amount of pixels for down sample skip
	//skipPx := downsample - 1
	fmt.Printf("Decoding section %d, %dx%d from y %d-%d\n", section, d.width, d.height, startY, endY)
	for y := startY; y < endY; y++ {
		for x := 0; x < d.width; x++ {
			// Do err and outside scan area logic
			err := readPixel(h, img, x, y)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func himawariDecode(sections []io.ReadSeekCloser) (*image.RGBA, error) {
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
	d := decodeMetadata(firstSection)
	//img = image.NewRGBA(image.Rect(0, 0, d.width, d.height*totalSections))
	// Continue to other sections
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		//decodeSection(firstSection, d, img)
		decodeToFile(firstSection, d)
	}()
	for section := 1; section < totalSections; section++ {
		wg.Add(1)
		// Decode data
		go func(f io.ReadSeeker) {
			defer wg.Done()
			h, err := DecodeFile(f)
			//err = decodeSection(h, d, img)
			decodeToFile(h, d)
			// TODO: err check
			if err != nil {
				//return nil, err
			}
		}(sections[section])
	}
	wg.Wait()

	fmt.Printf("Decoding done for %d sections\n", totalSections)
	return img, nil
}

func decodeMetadata(h *HMFile) sectionDecode {
	d := sectionDecode{
		width:  int(h.DataInfo.NumberOfColumns),
		height: int(h.DataInfo.NumberOfLines),
	}

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
	// different bands has different number of pixels bits, e.g., band 03 has 11
	coef := float64(data) / (math.Pow(2., float64(h.CalibrationInfo.ValidNumberOfBitsPerPixel)) - 2.)
	pc := pixel(coef, 1)
	img.Set(x, y, color.RGBA{R: uint8(pc), G: uint8(pc), B: uint8(pc), A: 255})

	return nil
}

// pixel Returns 255*coef clamping at coef, brightness adjusted
func pixel(coef, brig float64) int {
	return int(math.Min(coef*255*brig, 255))
}
