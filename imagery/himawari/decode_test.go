package main

import (
	"encoding/binary"
	"github.com/google/go-cmp/cmp"
	"io"
	"math"
	"os"
	"strings"
	"testing"
	"unicode"
)

func TestDecodeMetadata(t *testing.T) {
	f, err := os.Open("test-data/HS_H09_20231031_1340_B02_FLDK_R10_S0110.DAT")
	if err != nil {
		t.Error(err)
	}
	hw, err := DecodeFile(f)
	if err != nil {
		t.Error(err)
	}
	diff := cmp.Diff(&HMFile{
		BasicInfo: BasicInformation{
			BlockNumber:          1,
			BlockLength:          282,
			TotalHeaderBlocks:    11,
			ByteOrder:            binary.LittleEndian,
			Satellite:            [16]byte(c("Himawari-9")),
			ProcessingCenter:     [16]byte(c("MSC")),
			ObservationArea:      [4]byte(c("FLDK")),
			ObservationAreaInfo:  [2]byte(c("RT")),
			ObservationTimeline:  1340,
			ObservationStartTime: 60248.56968491159,
			ObservationEndTime:   60248.57007103656,
			FileCreationTime:     60248.57473379629,
			TotalHeaderLength:    1523,
			TotalDataLength:      24200000,
			QualityFlag1:         0,
			QualityFlag2:         0,
			QualityFlag3:         77,
			QualityFlag4:         1,
			FileFormatVersion:    [32]byte(c("1.3")),
			FileName:             [128]byte(c("HS_H09_20231031_1340_B02_FLDK_R10_S0110.DAT")),
			Spare:                [40]byte{},
		},
		DataInfo: DataInformationBlock{
			BlockNumber:          2,
			BlockLength:          50,
			NumberOfBitsPerPixel: 16,
			NumberOfColumns:      11000,
			NumberOfLines:        1100,
			CompressionFlag:      0,
			Spare:                [40]byte{},
		},
		ProjectionInfo: ProjectionInformationBlock{
			BlockNumber:             3,
			BlockLength:             127,
			SubLon:                  140.7,
			CFAC:                    40932549,
			LFAC:                    40932549,
			COFF:                    5500.5,
			LOFF:                    5500.5,
			DistanceFromEarthCenter: 42164,
			EarthEquatorialRadius:   6378.137,
			EarthPolarRadius:        6356.7523,
			RatioDiff:               0.0066943844,
			RatioPolar:              0.993305616,
			RatioEquatorial:         1.006739501,
			SDCoefficient:           1737122264,
			ResamplingSize:          4,
			Spare:                   [40]byte{},
		},
		NavigationInfo: NavigationInformationBlock{
			BlockNumber:                  4,
			BlockLength:                  139,
			NavigationTime:               60248.56964875857,
			SSPLongitude:                 140.7714266029078,
			SSPLatitude:                  0.0005093745779982038,
			DistanceFromEarthToSatellite: 42167.43974992831,
			NadirLongitude:               140.70940097119754,
			NadirLatitude:                -0.18059667277279684,
			SunPosition: Position{
				X: -117768009.68104868,
				Y: -83032296.6181277,
				Z: -35994351.16214465,
			},
			MoonPosition: Position{
				X: 115780.14199159983,
				Y: 323216.6892999755,
				Z: 168523.58232513716,
			},
			Spare: [40]byte{},
		},
		CalibrationInfo: CalibrationInformationBlock{
			BlockNumber:                       5,
			BlockLength:                       147,
			BandNumber:                        2,
			CentralWaveLength:                 0.509930,
			ValidNumberOfBitsPerPixel:         11,
			CountValueOfErrorPixels:           65535,
			CountValueOfPixelsOutsideScanArea: 65534,
			SlopeForCountRadianceEq:           0.35414147058823525,
			InterceptForCountRadianceEq:       -7.082829411764705,
			Infrared:                          InfraredBand{},
			Visible: VisibleBand{
				Albedo:              0.00166101782189072,
				UpdateTime:          57822.000000,
				CalibratedSlope:     0.354141470588,
				CalibratedIntercept: -7.082829411765,
				Spare:               [80]byte{},
			},
		},
		InterCalibrationInfo: InterCalibrationInformationBlock{
			BlockNumber:                6,
			BlockLength:                259,
			GSICSIntercept:             -10000000000.000000,
			GSICSSlope:                 -10000000000.000000,
			GSICSQuadratic:             -10000000000.000000,
			RadianceBias:               -10000000000.000000,
			RadianceUncertainty:        -10000000000.000000,
			RadianceStandardScene:      -10000000000.000000,
			GSICSCorrectionStart:       -10000000000.000000,
			GSICSCorrectionEnd:         -10000000000.000000,
			GSICSCalibrationUpperLimit: -10000000000.000000,
			GSICSCalibrationLowerLimit: -10000000000.000000,
			GSICSFileName:              [128]byte{},
			Spare:                      [56]byte{},
		},
		SegmentInfo: SegmentInformationBlock{
			BlockNumber:                   7,
			BlockLength:                   47,
			SegmentTotalNumber:            10,
			SegmentSequenceNumber:         1,
			FirstLineNumberOfImageSegment: 1,
			Spare:                         [40]byte{},
		},
		NavigationCorrectionInfo: NavigationCorrectionInformationBlock{
			BlockNumber:                  8,
			BlockLength:                  81,
			CenterColumnOfRotation:       1,
			CenterLineOfRotation:         1,
			AmountOfRotationalCorrection: 0,
			NumberOfCorrectionInfo:       2,
			Corrections: []NavigationCorrection{
				{
					LineNumberAfterRotation:        1,
					ShiftAmountForColumnCorrection: 0,
					ShiftAmountForLineCorrection:   0,
				},
				{
					LineNumberAfterRotation:        1100,
					ShiftAmountForColumnCorrection: 0,
					ShiftAmountForLineCorrection:   0,
				},
			},
			Spare: [40]byte{},
		},
		ObservationTimeInfo: ObservationTimeInformationBlock{
			BlockNumber:              9,
			BlockLength:              85,
			NumberOfObservationTimes: 4,
			Observations: []ObservationTime{
				{
					LineNumber:      1,
					ObservationTime: 60248.56968491159,
				},
				{
					LineNumber:      383,
					ObservationTime: 60248.56988264495,
				},
				{
					LineNumber:      875,
					ObservationTime: 60248.57007103656,
				},
				{
					LineNumber:      1100,
					ObservationTime: 60248.57007103656,
				},
			},
			Spare: [40]byte{},
		},
		ErrorInfo: ErrorInformationBlock{
			BlockNumber:    10,
			BlockLength:    47,
			NumberOfErrors: 0,
			Errors:         make([]ErrorInformation, 0),
			Spare:          [40]byte{},
		},
		SpareInfo: SpareInformationBlock{
			BlockNumber: 11,
			BlockLength: 259,
			Spare:       [256]byte{},
		},
	},
		hw,
		cmp.FilterPath(func(p cmp.Path) bool {
			field := p.Last().String()
			return field == ".ImageData" || startsWithLowercase(strings.Replace(field, ".", "", 1))
		}, cmp.Ignore()),
		cmp.AllowUnexported(HMFile{}),
	)
	if diff != "" {
		t.Errorf("received and expected not equal: %s", diff)
	}
}

func TestReadData(t *testing.T) {
	//f, err := os.Open("test-data/HS_H09_20231031_1340_B02_FLDK_R10_S0110.DAT.bz2")
	//if err != nil {
	//	t.Error(err)
	//}
	//hw, err := DecodeFile(bzip2.NewReader(f))
	//desiredCount := 11000 * 1100
}

func TestReadPixel(t *testing.T) {
	f, err := os.Open("test-data/HS_H09_20231031_1340_B02_FLDK_R10_S0110.DAT")
	if err != nil {
		t.Error(err)
	}
	hw, err := DecodeFile(f)
	px, _ := hw.ReadPixel()
	if px != (hw.CalibrationInfo.CountValueOfPixelsOutsideScanArea) {
		t.Errorf("expected %d but got %d for first pixel", hw.CalibrationInfo.CountValueOfPixelsOutsideScanArea, px)
	}
	count := 1
	desiredCount := 11000 * 1100
	for {
		_, err = hw.ReadPixel()
		if err == io.EOF {
			break
		}
		count++
	}

	if count != desiredCount {
		t.Errorf("expected to read %d pixels but read %d", desiredCount, count)
	}
}

func BenchmarkHMFile_ReadPixel(b *testing.B) {
	f, err := os.Open("test-data/HS_H09_20231031_1340_B02_FLDK_R10_S0110.DAT")
	if err != nil {
		b.Error(err)
	}
	hw, err := DecodeFile(f)
	px, _ := hw.ReadPixel()
	if px != (hw.CalibrationInfo.CountValueOfPixelsOutsideScanArea) {
		b.Errorf("expected %d but got %d for first pixel", hw.CalibrationInfo.CountValueOfPixelsOutsideScanArea, px)
	}
	count := 1
	for {
		if count > b.N {
			break
		}
		_, err = hw.ReadPixel()
		if err == io.EOF {
			break
		}
		count++
	}
}

func TestReadSkipSinglePixel(t *testing.T) {
	f, err := os.Open("test-data/HS_H09_20231031_1340_B02_FLDK_R10_S0110.DAT")
	if err != nil {
		t.Error(err)
	}
	hw, err := DecodeFile(f)
	totalPixels := 11000 * 1100
	skip := totalPixels - 1
	desiredCount := totalPixels - skip
	_, _ = hw.ReadPixel()
	err = hw.Skip(skip)

	count := 1
	if err != nil {
		t.Errorf("failed to skip %s", err)
	}
	for {
		_, err = hw.ReadPixel()
		if err == io.EOF {
			break
		} else if err != nil {
			t.Errorf("failed to read pixel: %s", err)
		}
		count++
	}

	if count != desiredCount {
		t.Errorf("expected to read %d pixels but read %d", desiredCount, count)
	}
}

func TestReadSkipEntireFile(t *testing.T) {
	f, err := os.Open("test-data/HS_H09_20231031_1340_B02_FLDK_R10_S0110.DAT")
	if err != nil {
		t.Error(err)
	}
	hw, err := DecodeFile(f)
	totalPixels := 11000 * 1100
	err = hw.Skip(totalPixels)

	count := 0
	if err != nil {
		t.Errorf("failed to skip %s", err)
	}
	for {
		_, err = hw.ReadPixel()
		if err == io.EOF {
			break
		} else if err != nil {
			t.Errorf("failed to read pixel: %s", err)
		}
		count++
	}

	if count != 0 {
		t.Errorf("expected to read %d pixels but read %d", 0, count)
	}
}

func FuzzHMFile_Skip(f *testing.F) {
	f.Add(5)
	f.Fuzz(func(t *testing.T, skip int) {
		skip = int(math.Abs(float64(skip))) * 1000
		f, err := os.Open("test-data/HS_H09_20231031_1340_B02_FLDK_R10_S0110.DAT")
		if err != nil {
			t.Error(err)
		}
		hw, err := DecodeFile(f)
		totalPixels := 11000 * 1100
		desiredCount := totalPixels - skip
		_, _ = hw.ReadPixel()
		err = hw.Skip(skip)

		count := 1
		if err != nil {
			t.Errorf("failed to skip %s", err)
		}
		for {
			_, err = hw.ReadPixel()
			if err == io.EOF {
				break
			} else if err != nil {
				t.Errorf("failed to read pixel: %s", err)
			}
			count++
		}

		if count != desiredCount {
			t.Errorf("expected to read %d pixels but read %d with skip value of %d", desiredCount, count, skip)
		}
	})
}

// c Util function to return a byte array of and padded
func c(s string) []byte {
	c := make([]byte, 1024)
	copy(c, s)
	return c
}

// startsWithLowercase checks if the string starts with a lowercase letter.
func startsWithLowercase(s string) bool {
	if s == "" {
		return false // An empty string does not start with a lowercase letter.
	}

	// Get the first rune (character) of the string.
	r := []rune(s)[0]

	// Check if the first rune is a lowercase letter.
	return unicode.IsLower(r)
}
