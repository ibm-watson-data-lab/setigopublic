package main

import (
  "time"
  "gopkg.in/guregu/null.v3"
)

type SignalDbRow struct {
	UniqueID   string          `json:"uniqueid"`
	Time       time.Time       `json:"time"`
	ActTyp     null.String  `json:"acttype"`
	TGTID      null.Int   `json:"tgtid"`
	Catalog    null.String  `json:"catalog"`
	Ra2000Hr   null.Float `json:"ra2000hr"`
	Dec2000Deg null.Float `json:"dec2000deg"`
	Power      null.Float `json:"power"`
	SNR        null.Float `json:"snr"`
	FreqMHZ    null.Float `json:"freqmhz"`
	DriftHZS   null.Float `json:"drifthzs"`
	WIDHZ      null.Float `json:"widhz"`
	POL        null.String  `json:"pol"`
	SigTyp     null.String  `json:"sigtyp"`
	PPeriods   null.Float `json:"pperiods"`
	NPul       null.Int   `json:"npul"`
	IntTimes   null.Float `json:"inttimes"`
	TSCPAZDEG  null.Float `json:"tscpazdeg"`
	TSCPELDEG  null.Float `json:"tscpeldeg"`
	BeamNo     null.Int   `json:"beamno"`
	SigClass   null.String  `json:"sigclass"`
	SigReason  null.String  `json:"sigreason"`
	CandReason null.String  `json:"candreason"`
}

type ArchiveCompampPath struct {
	UniqueID        string         `json:"uniqueid"`
	Container       string         `json:"container"`
  ObjectName      string         `json:"objectname"`
  Created_TS      time.Time         `json:"aca_created_ts"`
  Last_Modified_TS time.Time        `json:"aca_last_modified_ts"`
}

type SignalDBJoinACAPath struct {
  SignalDbRow
  Container       string         `json:"container"`
  ObjectName      string         `json:"objectname"`
}

type CelestialCoordinates struct {
	RA  float64
	Dec float64
}

type KnownACACoordinate struct {
    RA2000HR float64  `json:"ra2000hr"`
    DEC2000DEG float64 `json:"dec2000deg"`
    NUMBER_OF_ACA int64 `json:"number_of_aca"`
  }