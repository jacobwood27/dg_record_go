package main

import (
	"time"
)

type throwstamp struct {
	num  int
	time time.Time
	disc string
	lat  float64
	lon  float64
}

type throwstamps []throwstamp
