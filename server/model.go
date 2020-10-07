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
	Name     string
	Ip       []string
	Port     []string
	StartAt  time.Time
	FinishAt time.Time
}

type Finger struct {
	Addr      string
	Port      string
	TaskID    string
	AssetID   string
	Apps      []string
	Detail    []scanner.MatchedApp
	CreatedAt time.Time
}
