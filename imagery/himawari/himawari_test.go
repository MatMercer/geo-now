package main

import (
	"bufio"
	"encoding/binary"
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
	b := buffer.New(16 * 1024 * 1024)
	r, pw := nio.Pipe(b)
	w := bufio.NewWriter(pw)
	defer pw.Close()

	// Open the file for writing
	file, err := os.Create(fmt.Sprintf("section_%d.bmp", h.SegmentInfo.SegmentSequenceNumber))
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

	// Bits per pixel
	bitsPerPixel := 8

	// Calculate row size and padding
	rowSize := ((bitsPerPixel*d.width + 31) / 32) * 4
	padding := rowSize - d.width

	// Write BMP header
	writeBMPHeader(w, d.width, d.height, bitsPerPixel)

	fmt.Printf("Decoding section %d, %dx%d from y %d-%d\n", section, d.width, d.height, startY, endY)
	var pair [2]byte
	for y := startY; y < endY; y++ {
		for x := 0; x < d.width; x++ {
			_, err := h.ImageData.Read(pair[:])
			if err != nil {
				return err
			}
			// TODO check endianess
			p := uint16(pair[0]) | uint16(pair[1])<<8

			// Do err and outside scan area logic
			if p == h.CalibrationInfo.CountValueOfPixelsOutsideScanArea || p == h.CalibrationInfo.CountValueOfErrorPixels {
				w.WriteByte(0)
			} else {
				// Get a number between 0 and 1 from max number of pixels
				// Different bands has different number of pixels bits, e.g., band 03 has 11
				coef := float64(p) / (math.Pow(2., float64(h.CalibrationInfo.ValidNumberOfBitsPerPixel)) - 2.)
				brig := 1.0
				finalPixel := byte(math.Min(coef*255*brig, 255))
				w.WriteByte(finalPixel)
			}

			// Pad row to multiple of 4 bytes
			for x := 0; x < padding; x++ {
				err = w.WriteByte(0)
				if err != nil {
					panic(err)
				}
			}
		}
	}
	return nil
}

func writeBMPHeader(w *bufio.Writer, width, height, bitsPerPixel int) {
	// Calculate row size and image size
	rowSize := ((bitsPerPixel*width + 31) / 32) * 4
	imageSize := rowSize * height
	fileSize := 14 + 40 + 1024 + imageSize // File header + Info header + Palette + Image data

	// File header (14 bytes)
	w.Write([]byte{'B', 'M'})                                // Signature
	binary.Write(w, binary.LittleEndian, uint32(fileSize))   // File size
	binary.Write(w, binary.LittleEndian, uint32(0))          // Reserved
	binary.Write(w, binary.LittleEndian, uint32(14+40+1024)) // Offset to pixel data

	// Image header (40 bytes)
	binary.Write(w, binary.LittleEndian, uint32(40))           // Header size
	binary.Write(w, binary.LittleEndian, int32(width))         // Image width
	binary.Write(w, binary.LittleEndian, int32(height))        // Image height
	binary.Write(w, binary.LittleEndian, uint16(1))            // Number of color planes
	binary.Write(w, binary.LittleEndian, uint16(bitsPerPixel)) // Bits per pixel
	binary.Write(w, binary.LittleEndian, uint32(0))            // No compression
	binary.Write(w, binary.LittleEndian, uint32(imageSize))    // Image size
	binary.Write(w, binary.LittleEndian, int32(2835))          // X pixels per meter (72 DPI)
	binary.Write(w, binary.LittleEndian, int32(2835))          // Y pixels per meter (72 DPI)
	binary.Write(w, binary.LittleEndian, uint32(256))          // Number of colors in palette
	binary.Write(w, binary.LittleEndian, uint32(0))            // All colors are important

	// Color palette (256 * 4 = 1024 bytes)
	for i := 0; i < 256; i++ {
		w.WriteByte(uint8(i)) // Blue
		w.WriteByte(uint8(i)) // Green
		w.WriteByte(uint8(i)) // Red
		w.WriteByte(0)        // Reserved
	}
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
	var data uint16
	err := h.ReadPixel(&data)
	if err != nil {
		return fmt.Errorf("failed to read pixel at %d:%d: %s", x, y, err)
	}
	if data == h.CalibrationInfo.CountValueOfPixelsOutsideScanArea || data == h.CalibrationInfo.CountValueOfErrorPixels {
		img.Set(x, y, color.Black)
		return nil
	}

	// Get a number between 0 and 1 from max number of pixels
	// different bands has different number of pixels bits, e.g., band 03 has 11
	brig := 1.0
	coef := float64(data) / (math.Pow(2., float64(h.CalibrationInfo.ValidNumberOfBitsPerPixel)) - 2.)
	pc := int(math.Min(coef*255*brig, 255))
	img.Set(x, y, color.RGBA{R: uint8(pc), G: uint8(pc), B: uint8(pc), A: 255})

	return nil
}

// pixel Returns 255*coef clamping at coef, brightness adjusted
func pixel(coef, brig float64) int {
	return int(math.Min(coef*255*brig, 255))
}
