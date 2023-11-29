package main

import (
	"github.com/google/go-cmp/cmp"
	"os"
	"testing"
)

func TestDecode(t *testing.T) {
	f, err := os.Open("sample-data/HS_H09_20231031_1340_B02_FLDK_R10_S0110.DAT")
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
			ByteOrder:            LittleEndian,
			Satellite:            [16]C(c("Himawari-9")),
			ProcessingCenter:     [16]C(c("MSC")),
			ObservationArea:      [4]C(c("FLDK")),
			ObservationAreaInfo:  [2]C(c("RT")),
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
			FileFormatVersion:    [32]C(c("1.3")),
			FileName:             [128]C(c("HS_H09_20231031_1340_B02_FLDK_R10_S0110.DAT")),
			Spare:                [40]C{},
		},
		DataInfo: DataInformationBlock{
			BlockNumber:          2,
			BlockLength:          50,
			NumberOfBitsPerPixel: 16,
			NumberOfColumns:      11000,
			NumberOfLines:        1100,
			CompressionFlag:      0,
			Spare:                [40]C{},
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
			Spare:                   [40]C{},
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
			Spare: [40]C{},
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
				Spare:               [80]C{},
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
			GSICSFileName:              [128]C{},
			Spare:                      [56]C{},
		},
		SegmentInfo: SegmentInformationBlock{
			BlockNumber:                   7,
			BlockLength:                   47,
			SegmentTotalNumber:            10,
			SegmentSequenceNumber:         1,
			FirstLineNumberOfImageSegment: 1,
			Spare:                         [40]C{},
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
			Spare: [40]C{},
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
			Spare: [40]C{},
		},
		ErrorInfo: ErrorInformationBlock{
			BlockNumber:    10,
			BlockLength:    47,
			NumberOfErrors: 0,
			Errors:         make([]ErrorInformation, 0),
			Spare:          [40]C{},
		},
	}, hw)
	if diff != "" {
		t.Errorf("received and expected not equal: %s", diff)
	}
}

func c(s string) []C {
	c := make([]C, 1024)
	copy(c, []C(s))
	return c
}
