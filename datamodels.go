package main

import "time"

type SignalDbRow struct {
  UniqueID string `json:"uniqueid"` 
  Time time.Time `json:"time"`
  ActTyp string `json:"acttype"`
  TGTID int `json:"tgtid"`
  Catalog string `json:"catalog"`
  Ra2000Hr float64 `json:"ra2000hr"`
  Dec2000Deg float64 `json:"dec2000deg"`
  Power float64 `json:"power"`
  SNR float64 `json:"snr"`
  FreqMHZ float64 `json:"freqmhz"`
  DriftHZS float64 `json:"drifthzs"`
  WIDHZ float64 `json:"widhz"`
  POL string `json:"pol"`
  SigType string `json:"sigtype"`
  PPeriods float64 `json:"pperiods"`
  NPul int `json:"npul"`
  IntTimes float64 `json:"inttimes"`
  TSCPAZDEG float64 `json:"tscpazdeg"`
  TSCPELDEG float64 `json:"tscpeldeg"`
  BeamNo int `json:"beamno"`
  SigClass string `json:"sigclass"`
  SigReason string `json:"sigreason"`
  CandReason string `json:"candreason"`
}

type SignalDbRows []SignalDbRow