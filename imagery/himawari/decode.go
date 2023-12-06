package main

import (
	"bytes"
	"encoding/binary"
	"io"
)

const (
	LittleEndian = 0
	BigEndian    = 1
)

type HMFile struct {
	BasicInfo                BasicInformation
	DataInfo                 DataInformationBlock
	ProjectionInfo           ProjectionInformationBlock
	NavigationInfo           NavigationInformationBlock
	CalibrationInfo          CalibrationInformationBlock
	InterCalibrationInfo     InterCalibrationInformationBlock
	SegmentInfo              SegmentInformationBlock
	NavigationCorrectionInfo NavigationCorrectionInformationBlock
	ObservationTimeInfo      ObservationTimeInformationBlock
	ErrorInfo                ErrorInformationBlock
	SpareInfo                SpareInformationBlock
	ImageData                []uint16
}

type Position struct {
	X float64
	Y float64
	Z float64
}

type BasicInformation struct {
	BlockNumber          uint8
	BlockLength          uint16
	TotalHeaderBlocks    uint16
	ByteOrder            uint8
	Satellite            [16]byte
	ProcessingCenter     [16]byte
	ObservationArea      [4]byte
	ObservationAreaInfo  [2]byte
	ObservationTimeline  uint16
	ObservationStartTime float64
	ObservationEndTime   float64
	FileCreationTime     float64
	TotalHeaderLength    uint32
	TotalDataLength      uint32
	QualityFlag1         uint8
	QualityFlag2         uint8
	QualityFlag3         uint8
	QualityFlag4         uint8
	FileFormatVersion    [32]byte
	FileName             [128]byte
	Spare                [40]byte
}

type DataInformationBlock struct {
	BlockNumber          uint8
	BlockLength          uint16
	NumberOfBitsPerPixel uint16
	NumberOfColumns      uint16
	NumberOfLines        uint16
	CompressionFlag      uint8
	Spare                [40]byte
}

type ProjectionInformationBlock struct {
	BlockNumber             uint8
	BlockLength             uint16
	SubLon                  float64
	CFAC                    uint32
	LFAC                    uint32
	COFF                    float32
	LOFF                    float32
	DistanceFromEarthCenter float64
	EarthEquatorialRadius   float64
	EarthPolarRadius        float64
	RatioDiff               float64
	RatioPolar              float64
	RatioEquatorial         float64
	SDCoefficient           float64
	ResamplingTypes         uint16
	ResamplingSize          uint16
	Spare                   [40]byte
}

type NavigationInformationBlock struct {
	BlockNumber                  uint8
	BlockLength                  uint16
	NavigationTime               float64
	SSPLongitude                 float64
	SSPLatitude                  float64
	DistanceFromEarthToSatellite float64
	NadirLongitude               float64
	NadirLatitude                float64
	SunPosition                  Position
	MoonPosition                 Position
	Spare                        [40]byte
}

type CalibrationInformationBlock struct {
	BlockNumber                       uint8
	BlockLength                       uint16
	BandNumber                        uint16
	CentralWaveLength                 float64
	ValidNumberOfBitsPerPixel         uint16
	CountValueOfErrorPixels           uint16
	CountValueOfPixelsOutsideScanArea uint16
	SlopeForCountRadianceEq           float64
	InterceptForCountRadianceEq       float64
	Infrared                          InfraredBand
	Visible                           VisibleBand
}

type InfraredBand struct {
	BrightnessTemp    float64
	BrightnessC1      float64
	BrightnessC2      float64
	Radiance          float64
	RadianceC1        float64
	RadianceC2        float64
	SpeedOfLight      float64
	PlanckConstant    float64
	BoltzmannConstant float64
	Spare             [40]byte
}

type VisibleBand struct {
	Albedo              float64
	UpdateTime          float64
	CalibratedSlope     float64
	CalibratedIntercept float64
	Spare               [80]byte
}

type InterCalibrationInformationBlock struct {
	BlockNumber                uint8
	BlockLength                uint16
	GSICSIntercept             float64
	GSICSSlope                 float64
	GSICSQuadratic             float64
	RadianceBias               float64
	RadianceUncertainty        float64
	RadianceStandardScene      float64
	GSICSCorrectionStart       float64
	GSICSCorrectionEnd         float64
	GSICSCalibrationUpperLimit float32
	GSICSCalibrationLowerLimit float32
	GSICSFileName              [128]byte
	Spare                      [56]byte
}

type SegmentInformationBlock struct {
	BlockNumber                   uint8
	BlockLength                   uint16
	SegmentTotalNumber            uint8
	SegmentSequenceNumber         uint8
	FirstLineNumberOfImageSegment uint16
	Spare                         [40]byte
}

type NavigationCorrectionInformationBlock struct {
	BlockNumber                  uint8
	BlockLength                  uint16
	CenterColumnOfRotation       float32
	CenterLineOfRotation         float32
	AmountOfRotationalCorrection float64
	NumberOfCorrectionInfo       uint16
	Corrections                  []NavigationCorrection
	Spare                        [40]byte
}

type NavigationCorrection struct {
	LineNumberAfterRotation        uint16
	ShiftAmountForColumnCorrection float32
	ShiftAmountForLineCorrection   float32
}

type ObservationTimeInformationBlock struct {
	BlockNumber              uint8
	BlockLength              uint16
	NumberOfObservationTimes uint16
	Observations             []ObservationTime
	Spare                    [40]byte
}

type ObservationTime struct {
	LineNumber      uint16
	ObservationTime float64
}

type ErrorInformationBlock struct {
	BlockNumber    uint8
	BlockLength    uint32
	NumberOfErrors uint16
	Errors         []ErrorInformation
	Spare          [40]byte
}

type ErrorInformation struct {
	LineNumber     uint16
	NumberOfPixels uint16
}

type SpareInformationBlock struct {
	BlockNumber uint8
	BlockLength uint16
	Spare       [256]byte
}

func DecodeFile(r io.Reader) (*HMFile, error) {
	// Decode basic info
	// uint8+uint16+uint16=5
	basicInfo := make([]byte, 5)
	_, err := r.Read(basicInfo)
	if err != nil {
		return nil, err
	}
	basicBuffer := bytes.NewBuffer(basicInfo)
	i := BasicInformation{}
	// Detect byte order
	read(r, binary.BigEndian, &i.ByteOrder)
	var o binary.ByteOrder
	if i.ByteOrder == LittleEndian {
		o = binary.LittleEndian
	} else {
		o = binary.BigEndian
	}
	// Read existing buffer
	read(basicBuffer, o, &i.BlockNumber)
	read(basicBuffer, o, &i.BlockLength)
	read(basicBuffer, o, &i.TotalHeaderBlocks)

	// Skip Byte order because already read and continue normal decoding
	read(r, o, &i.Satellite)
	read(r, o, &i.ProcessingCenter)
	read(r, o, &i.ObservationArea)
	read(r, o, &i.ObservationAreaInfo)
	read(r, o, &i.ObservationTimeline)
	read(r, o, &i.ObservationStartTime)
	read(r, o, &i.ObservationEndTime)
	read(r, o, &i.FileCreationTime)
	read(r, o, &i.TotalHeaderLength)
	read(r, o, &i.TotalDataLength)
	read(r, o, &i.QualityFlag1)
	read(r, o, &i.QualityFlag2)
	read(r, o, &i.QualityFlag3)
	read(r, o, &i.QualityFlag4)
	read(r, o, &i.FileFormatVersion)
	read(r, o, &i.FileName)
	read(r, o, &i.Spare)

	// Decode data information block
	d := DataInformationBlock{}
	read(r, o, &d.BlockNumber)
	read(r, o, &d.BlockLength)
	read(r, o, &d.NumberOfBitsPerPixel)
	read(r, o, &d.NumberOfColumns)
	read(r, o, &d.NumberOfLines)
	read(r, o, &d.CompressionFlag)
	read(r, o, &d.Spare)

	// Decode projection information block
	p := ProjectionInformationBlock{}
	read(r, o, &p.BlockNumber)
	read(r, o, &p.BlockLength)
	read(r, o, &p.SubLon)
	read(r, o, &p.CFAC)
	read(r, o, &p.LFAC)
	read(r, o, &p.COFF)
	read(r, o, &p.LOFF)
	read(r, o, &p.DistanceFromEarthCenter)
	read(r, o, &p.EarthEquatorialRadius)
	read(r, o, &p.EarthPolarRadius)
	read(r, o, &p.RatioDiff)
	read(r, o, &p.RatioPolar)
	read(r, o, &p.RatioEquatorial)
	read(r, o, &p.SDCoefficient)
	read(r, o, &p.ResamplingTypes)
	read(r, o, &p.ResamplingSize)
	read(r, o, &d.Spare)

	// Decode navigation information block
	n := NavigationInformationBlock{}
	read(r, o, &n.BlockNumber)
	read(r, o, &n.BlockLength)
	read(r, o, &n.NavigationTime)
	read(r, o, &n.SSPLongitude)
	read(r, o, &n.SSPLatitude)
	read(r, o, &n.DistanceFromEarthToSatellite)
	read(r, o, &n.NadirLongitude)
	read(r, o, &n.NadirLatitude)
	read(r, o, &n.SunPosition.X)
	read(r, o, &n.SunPosition.Y)
	read(r, o, &n.SunPosition.Z)
	read(r, o, &n.MoonPosition.X)
	read(r, o, &n.MoonPosition.Y)
	read(r, o, &n.MoonPosition.Z)
	read(r, o, &n.Spare)

	// Decode calibration info block
	c := CalibrationInformationBlock{}
	read(r, o, &c.BlockNumber)
	read(r, o, &c.BlockLength)
	read(r, o, &c.BandNumber)
	read(r, o, &c.CentralWaveLength)
	read(r, o, &c.ValidNumberOfBitsPerPixel)
	read(r, o, &c.CountValueOfErrorPixels)
	read(r, o, &c.CountValueOfPixelsOutsideScanArea)
	read(r, o, &c.SlopeForCountRadianceEq)
	read(r, o, &c.InterceptForCountRadianceEq)
	// Visible light
	if c.BandNumber < 7 {
		read(r, o, &c.Visible.Albedo)
		read(r, o, &c.Visible.UpdateTime)
		read(r, o, &c.Visible.CalibratedSlope)
		read(r, o, &c.Visible.CalibratedIntercept)
		read(r, o, &c.Visible.Spare)
	} else {
		// TODO: infrared, 112 means what is the end of the block
		read(r, o, make([]byte, 112))
	}

	// Decode inter calibration info block
	ci := InterCalibrationInformationBlock{}
	read(r, o, &ci.BlockNumber)
	read(r, o, &ci.BlockLength)
	read(r, o, &ci.GSICSIntercept)
	read(r, o, &ci.GSICSSlope)
	read(r, o, &ci.GSICSQuadratic)
	read(r, o, &ci.RadianceBias)
	read(r, o, &ci.RadianceUncertainty)
	read(r, o, &ci.RadianceStandardScene)
	read(r, o, &ci.GSICSCorrectionStart)
	read(r, o, &ci.GSICSCorrectionEnd)
	read(r, o, &ci.GSICSCalibrationUpperLimit)
	read(r, o, &ci.GSICSCalibrationLowerLimit)
	read(r, o, &ci.GSICSFileName)
	read(r, o, &ci.Spare)

	// Decode segment info block
	s := SegmentInformationBlock{}
	read(r, o, &s.BlockNumber)
	read(r, o, &s.BlockLength)
	read(r, o, &s.SegmentTotalNumber)
	read(r, o, &s.SegmentSequenceNumber)
	read(r, o, &s.FirstLineNumberOfImageSegment)
	read(r, o, &s.Spare)

	// Decode navigation correction block
	nc := NavigationCorrectionInformationBlock{}
	read(r, o, &nc.BlockNumber)
	read(r, o, &nc.BlockLength)
	read(r, o, &nc.CenterColumnOfRotation)
	read(r, o, &nc.CenterLineOfRotation)
	read(r, o, &nc.AmountOfRotationalCorrection)
	read(r, o, &nc.NumberOfCorrectionInfo)
	nc.Corrections = make([]NavigationCorrection, nc.NumberOfCorrectionInfo)
	for i := uint16(0); i < nc.NumberOfCorrectionInfo; i++ {
		correct := NavigationCorrection{}
		read(r, o, &correct.LineNumberAfterRotation)
		read(r, o, &correct.ShiftAmountForColumnCorrection)
		read(r, o, &correct.ShiftAmountForLineCorrection)
		nc.Corrections[i] = correct
	}
	read(r, o, &nc.Spare)

	// Decode observation time block
	ob := ObservationTimeInformationBlock{}
	read(r, o, &ob.BlockNumber)
	read(r, o, &ob.BlockLength)
	read(r, o, &ob.NumberOfObservationTimes)
	ob.Observations = make([]ObservationTime, ob.NumberOfObservationTimes)
	for i := uint16(0); i < ob.NumberOfObservationTimes; i++ {
		observation := ObservationTime{}
		read(r, o, &observation.LineNumber)
		read(r, o, &observation.ObservationTime)
		ob.Observations[i] = observation
	}
	read(r, o, &ob.Spare)

	// Decode error information block
	ei := ErrorInformationBlock{}
	read(r, o, &ei.BlockNumber)
	read(r, o, &ei.BlockLength)
	read(r, o, &ei.NumberOfErrors)
	ei.Errors = make([]ErrorInformation, ei.NumberOfErrors)
	for i := uint16(0); i < ei.NumberOfErrors; i++ {
		errorInfo := ErrorInformation{}
		read(r, o, &errorInfo.LineNumber)
		read(r, o, &errorInfo.NumberOfPixels)
		ei.Errors[i] = errorInfo
	}
	read(r, o, &ei.Spare)

	// Decode spare information block
	sp := SpareInformationBlock{}
	read(r, o, &sp.BlockNumber)
	read(r, o, &sp.BlockLength)
	read(r, o, &sp.Spare)

	// Decode data
	h := &HMFile{
		BasicInfo:                i,
		DataInfo:                 d,
		ProjectionInfo:           p,
		NavigationInfo:           n,
		CalibrationInfo:          c,
		InterCalibrationInfo:     ci,
		SegmentInfo:              s,
		NavigationCorrectionInfo: nc,
		ObservationTimeInfo:      ob,
		ErrorInfo:                ei,
		SpareInfo:                sp,
	}

	h.ImageData = make([]uint16, int(h.DataInfo.NumberOfColumns)*int(h.DataInfo.NumberOfLines))
	read(r, o, &h.ImageData)

	return h, nil
}

// read util function that reads and ignore error
func read(f io.Reader, o binary.ByteOrder, dst any) {
	_ = binary.Read(f, o, dst)
}

func (f *HMFile) ReadPixel() uint16 {
	return uint16(0)
}
