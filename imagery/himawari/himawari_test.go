package main

import (
	"bufio"
	"encoding/binary"
	"github.com/djherbis/buffer"
	"github.com/djherbis/nio"
	"matbm.net/geonow/imagery/colometry"
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

func TestDecodeMultiBand(t *testing.T) {
	// Profiling with pprof
	var wg sync.WaitGroup
	go func() {
		fmt.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	wg.Add(1)

	// B01 - 470
	// B02 - 510
	// B03 - 640
	var bandsToDecode = []int{1, 2, 3}

	files := make(map[int][]io.ReadSeekCloser)

	for _, band := range bandsToDecode {
		src := fmt.Sprintf("HS_H09_20231130_0030_B%02d_FLDK", band)
		dir := "./sample-data"
		sections, err := openFiles(dir, src)
		if err != nil {
			fmt.Printf("Failed to open himawari sections: %s\n", err)
			return
		}

		files[band] = sections
	}

	_, err := himawariDecodeMultiband(files)
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

func TestDecode(t *testing.T) {
	// Profiling with pprof
	var wg sync.WaitGroup
	go func() {
		fmt.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	wg.Add(1)

	// B01 - 470
	// B02 - 510
	// B03 - 640

	src := "HS_H09_20231130_0030_B01_FLDK_R10"
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

type HMDecode struct {
	*HMFile
	colR float64
	colG float64
	colB float64
}

func (h *HMDecode) Init() error {
	waveLength, err := h.GetWaveLength()
	colR, colG, colB, err := colometry.ToRGB(waveLength)
	if err != nil {
		return err
	}

	h.colR = colR
	h.colG = colG
	h.colB = colB

	return nil
}

func (h *HMDecode) DecimateCols(pair [2]byte) {
	// Decimate the columns
	for i := 0; i < h.DecodeInstructions.Decimate-1; i++ {
		_, _ = h.ImageData.Read(pair[:])
	}
}

func (h *HMDecode) DecimateLines() {
	// Decimate the lines
	for i := 0; i < h.DecodeInstructions.Decimate-1; i++ {
		io.CopyN(io.Discard, h.ImageData, int64(2*h.DataInfo.NumberOfColumns))
	}
}

func (h *HMDecode) NextPixel(pair [2]byte) (r, g, b float64, err error) {
	_, err = h.ImageData.Read(pair[:])
	if err != nil {
		return 0., 0., 0., err
	}
	// TODO check endianess
	p := uint16(pair[0]) | uint16(pair[1])<<8

	// Do err and outside scan area logic
	if p == h.CalibrationInfo.CountValueOfPixelsOutsideScanArea || p == h.CalibrationInfo.CountValueOfErrorPixels {
		return 0., 0., 0., nil
	} else {
		// Get a number between 0 and 1 from max number of pixels
		// Different bands has different number of pixels bits, e.g., band 03 has 11
		coef := float64(p) / (math.Pow(2., float64(h.CalibrationInfo.ValidNumberOfBitsPerPixel)) - 2.)
		brig := 1.0
		finalR := h.colR * coef * brig
		finalG := h.colG * coef * brig
		finalB := h.colB * coef * brig

		// Bitmaps uses BGR
		return finalR, finalG, finalB, nil
	}
}

// Decodes multiple bands of a section
func decodeToFileMultiband(files []*HMDecode) error {
	b := buffer.New(16 * 1024 * 1024)
	r, pw := nio.Pipe(b)
	w := bufio.NewWriter(pw)
	defer pw.Close()

	// Open the file for writing
	file, err := os.Create(fmt.Sprintf("section_%d.bmp", files[0].SegmentInfo.SegmentSequenceNumber))
	if err != nil {
		return err
	}
	defer file.Close()

	// Start a goroutine to write to the file
	go func() {
		_, _ = io.Copy(file, r)
	}()

	// Get the decode info
	d := files[0].DecodeInstructions

	// Start and End Y are the relative positions for the final image based in a section
	section := files[0].SegmentInfo.SegmentSequenceNumber
	startY := d.TargetHeight * int(section-1)
	endY := startY + d.TargetHeight

	// Bits per pixel
	bitsPerPixel := 24

	// Write BMP header
	writeBMPHeader(w, d.TargetWidth, d.TargetHeight, bitsPerPixel)

	fmt.Printf("Decoding section %d, %dx%d from y %d-%d with %d bands\n", section, d.TargetWidth, d.TargetHeight, startY, endY, len(files))
	var pair [2]byte

	for y := startY; y < endY; y++ {
		for x := 0; x < d.TargetWidth; x++ {
			var finalR, finalG, finalB float64
			for _, h := range files {
				r, g, b, err := h.NextPixel(pair)
				if err != nil {
					return err
				}

				finalR += r
				finalG += g
				finalB += b

				// Decimate the columns
				h.DecimateCols(pair)
			}

			finalR /= float64(len(files))
			finalG /= float64(len(files))
			finalB /= float64(len(files))

			// Bitmaps uses BGR
			pixel := []byte{byte(math.Min(finalB*255, 255)), byte(math.Min(finalG*255, 255)), byte(math.Min(finalR*255, 255))}
			w.Write(pixel)
		}
		// Decimate the lines
		for _, h := range files {
			h.DecimateLines()
		}
	}

	fmt.Printf("Decoding of section %d done\n", section)
	return nil
}

func decodeToFile(h *HMFile) error {
	// 16mb
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

	// Get the decode info
	d := h.DecodeInstructions

	// Start and End Y are the relative positions for the final image based in a section
	section := h.SegmentInfo.SegmentSequenceNumber
	startY := d.TargetHeight * int(section-1)
	endY := startY + d.TargetHeight

	// Bits per pixel
	bitsPerPixel := 24

	// Write BMP header
	writeBMPHeader(w, d.TargetWidth, d.TargetHeight, bitsPerPixel)

	fmt.Printf("Decoding section %d, %dx%d from y %d-%d\n", section, d.TargetWidth, d.TargetHeight, startY, endY)
	var pair [2]byte

	waveLength, err := h.GetWaveLength()
	colR, colG, colB, err := colometry.ToRGB(waveLength)
	if err != nil {
		return err
	}

	blackPixel := []byte{0, 0, 0}
	var outPixel []byte

	for y := startY; y < endY; y++ {
		for x := 0; x < d.TargetWidth; x++ {
			_, err := h.ImageData.Read(pair[:])
			if err != nil {
				return err
			}
			// TODO check endianess
			p := uint16(pair[0]) | uint16(pair[1])<<8

			// Do err and outside scan area logic
			if p == h.CalibrationInfo.CountValueOfPixelsOutsideScanArea || p == h.CalibrationInfo.CountValueOfErrorPixels {
				w.Write(blackPixel)
			} else {
				// Get a number between 0 and 1 from max number of pixels
				// Different bands has different number of pixels bits, e.g., band 03 has 11
				coef := float64(p) / (math.Pow(2., float64(h.CalibrationInfo.ValidNumberOfBitsPerPixel)) - 2.)
				brig := 1.0
				finalR := byte(math.Min(colR*coef*255*brig, 255))
				finalG := byte(math.Min(colG*coef*255*brig, 255))
				finalB := byte(math.Min(colB*coef*255*brig, 255))

				// Bitmaps uses BGR
				outPixel = []byte{finalB, finalG, finalR}

				w.Write(outPixel)
			}

			// Decimate the columns
			for i := 0; i < h.DecodeInstructions.Decimate-1; i++ {
				_, err = h.ImageData.Read(pair[:])
			}
		}
		// Decimate the lines
		for i := 0; i < h.DecodeInstructions.Decimate-1; i++ {
			io.CopyN(io.Discard, h.ImageData, int64(2*h.DataInfo.NumberOfColumns))
		}
	}
	return nil
}

func writeBMPHeader(w *bufio.Writer, width, height, bitsPerPixel int) {
	// Calculate row size and image size
	rowSize := ((bitsPerPixel*width + 31) / 32) * 4
	imageSize := rowSize * height
	fileSize := 14 + 40 + imageSize // File header + Info header + Image data

	// File header (14 bytes)
	w.Write([]byte{'B', 'M'})                              // Signature
	binary.Write(w, binary.LittleEndian, uint32(fileSize)) // File size
	binary.Write(w, binary.LittleEndian, uint32(0))        // Reserved
	binary.Write(w, binary.LittleEndian, uint32(14+40))    // Offset to pixel data

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
	binary.Write(w, binary.LittleEndian, uint32(0))            // Number of colors in palette
	binary.Write(w, binary.LittleEndian, uint32(0))            // All colors are important
}

// decodeSection deprecated
func decodeSection(h *HMFile, img *image.RGBA) error {
	// Start and End Y are the relative positions for the final image based in a section
	section := h.SegmentInfo.SegmentSequenceNumber
	d := h.DecodeInstructions
	startY := d.TargetHeight * int(section-1)
	endY := startY + d.TargetWidth
	// Amount of pixels for down sample skip
	//skipPx := downsample - 1
	fmt.Printf("Decoding section %d, %dx%d from y %d-%d\n", section, d.TargetWidth, d.TargetHeight, startY, endY)
	for y := startY; y < endY; y++ {
		for x := 0; x < d.TargetWidth; x++ {
			// Do err and outside scan area logic
			err := readPixel(h, img, x, y)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func himawariDecodeMultiband(bands map[int][]io.ReadSeekCloser) (*image.RGBA, error) {
	// 4Mb
	buffSize := 4 * 1024 * 1024

	// Close all files
	defer func() {
		for _, sections := range bands {
			for _, s := range sections {
				_ = s.Close()
			}
		}
	}()

	// Map between section number and himawari file (for visible light it's 3)
	himawariFiles := make(map[int][]*HMFile)
	for _, sections := range bands {
		for _, s := range sections {
			hw, err := DecodeFile(s, buffSize)
			if err != nil {
				return nil, err
			}
			himawariFiles[int(hw.SegmentInfo.SegmentSequenceNumber)] = append(himawariFiles[int(hw.SegmentInfo.SegmentSequenceNumber)], hw)
		}
	}

	var img *image.RGBA

	var wg sync.WaitGroup
	// TODO pass section number to decode to make life easier later
	for _, files := range himawariFiles {
		wg.Add(1)
		go func(files []*HMFile) {
			defer wg.Done()

			// Create a list of files to decode and initialize metadata
			decodes := make([]*HMDecode, len(files))
			for i, f := range files {
				decodes[i] = &HMDecode{HMFile: f}
				err := decodes[i].Init()
				if err != nil {
					fmt.Printf("Failed to init decode: %s\n", err)
					return
				}
			}

			decodeToFileMultiband(decodes)
		}(files)
	}
	wg.Wait()

	fmt.Printf("Decoding done for %d sections\n", len(himawariFiles))
	return img, nil
}

func himawariDecode(sections []io.ReadSeekCloser) (*image.RGBA, error) {
	// 16mb
	buffSize := 2 * 1024 * 1024

	defer func() {
		for _, s := range sections {
			_ = s.Close()
		}
	}()
	var img *image.RGBA

	// Decode first section to gather file info
	firstSection, err := DecodeFile(sections[0], buffSize)
	if err != nil {
		return nil, fmt.Errorf("failed to decode first section: %s", err)
	}
	totalSections := len(sections)
	//d := decodeMetadata(firstSection)
	//img = image.NewRGBA(image.Rect(0, 0, d.width, d.height*totalSections))
	// Continue to other sections
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		//decodeSection(firstSection, d, img)
		decodeToFile(firstSection)
	}()
	for section := 1; section < totalSections; section++ {
		wg.Add(1)
		// Decode data
		go func(f io.ReadSeeker) {
			defer wg.Done()
			h, err := DecodeFile(f, buffSize)
			//err = decodeSection(h, d, img)
			decodeToFile(h)
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

// decodeMetadata deprecated
func decodeMetadata(h *HMFile) sectionDecode {
	d := sectionDecode{
		width:  int(h.DataInfo.NumberOfColumns),
		height: int(h.DataInfo.NumberOfLines),
	}

	return d
}

// deprecated
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

// pixel deprecated Returns 255*coef clamping at coef, brightness adjusted
func pixel(coef, brig float64) int {
	return int(math.Min(coef*255*brig, 255))
}
