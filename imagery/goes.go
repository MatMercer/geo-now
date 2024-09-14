package imagery

import (
	"bufio"
	"github.com/davidbyttow/govips/v2/vips"
	"io"
	"log"
	"net/http"
)

const (
	host         = "https://cdn.star.nesdis.noaa.gov"
	eastGeocolor = "/GOES16/ABI/FD/GEOCOLOR/latest.jpg"
)

type GoesSource struct {
	MaxWidth int
}

// DownloadImage downloads a goes-east, full disk, geo-color, latest image
func (GoesSource) DownloadImage() (*bufio.Reader, error) {
	client := http.Client{}
	resp, err := client.Get(host + eastGeocolor)
	if err != nil {
		return nil, err
	}

	return bufio.NewReader(resp.Body), nil
}

// PostProcess Crop the top/bottom 16px of the GOES image since they are unnecessary
func (g GoesSource) PostProcess(src io.Reader, dst io.Writer) error {
	img, err := vips.NewImageFromReader(src)
	if err != nil {
		return err
	}
	err = img.ExtractArea(0, 16, img.Width(), img.Height()-32)
	if err != nil {
		return err
	}
	ratio := float64(img.Width()) / float64(img.Height())
	err = img.Thumbnail(g.MaxWidth, int(float64(g.MaxWidth)*ratio), vips.InterestingNone)
	if err != nil {
		return err
	}
	jpeg, metadata, err := img.ExportJpeg(nil)

	n, err := dst.Write(jpeg)
	if err != nil {
		return err
	}
	log.Printf("Goes post process: %dx%d %d bytes", metadata.Width, metadata.Height, n)

	return nil
}

func (g GoesSource) SourceURL() string {
	return host + eastGeocolor
}
