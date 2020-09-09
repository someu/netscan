package main

import (
	"fakescan/scanner"
	"time"
)

type Asset struct {
	Address   string
	Fingers   []*scanner.MatchedApp
	CreatedAt time.Time
}

type Scan struct {
	Target   []string
	StartAt  time.Time
	FinishAt time.Time
}
