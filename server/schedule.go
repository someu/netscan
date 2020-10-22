package main

import (
	"context"
	"netscan/appscan"
	"time"
)

type Scheduler struct {
	scanner *appscan.Scanner
}

func (s *Scheduler) SetConcurrent(c int) int {
	if c > 0 {
		s.scanner.SetConcurrent(c)
	}
	return s.scanner.Concurrent
}

func (s *Scheduler) CreateScan(scan Scan) {
	s.scanner.Scan(scan.Ip, scan.Port, func(result *appscan.MatchedResult) {
		var apps []string
		for _, app := range result.Apps {
			apps = append(apps, app.Name)
		}
		finger := Finger{
			Addr:      result.Url,
			Apps:      apps,
			Detail:    result.Apps,
			CreatedAt: time.Now(),
		}
		log.Printf("pre insertss")
		if _, err := fingerCollection.InsertOne(context.TODO(), finger); err != nil {
			log.Printf("Insert finger failed, %s", err.Error())
		} else {
			log.Printf("Insert finger success")
		}
	})
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		scanner: scanner.NewScanner(),
	}
}
