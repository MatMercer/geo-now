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
