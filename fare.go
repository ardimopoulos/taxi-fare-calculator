package taxifarecalculator

import (
	"math"
	"sync"
	"time"
)

type Fare struct {
	id   int
	fare float64
}

const (
	flagAmount      = 1.30
	minimumAmount   = 3.47
	idleAmount      = 11.90
	singleAmount    = 0.74
	increasedAmount = 1.30
)

func NewFare(id int, fare float64) *Fare {
	return &Fare{id, fare}
}

func (f *Fare) GetId() int {
	return f.id
}

func (f *Fare) GetFare() float64 {
	return f.fare
}

func CalculateFare(batch <-chan []*Position, wg *sync.WaitGroup) chan *Fare {
	var p1 *Position
	rideFare := make(chan *Fare)

	go func() {
		var id int

		for pos := range batch {
			id++
			fare := flagAmount

			for i, p := range pos {
				if i == 0 {
					p1 = p
					continue
				}

				if p1 == nil {
					continue
				}

				fare += CalculateSegmentFare(p1, p)
				p1 = p
			}

			if fare < minimumAmount {
				fare = minimumAmount
			}

			rideFare <- NewFare(id, fare)
		}

		close(rideFare)
		wg.Done()
	}()

	return rideFare
}

func CalculateSegmentFare(p1, p2 *Position) float64 {
	d := p1.CalculateHaversineDistance(p2)
	t := p1.CalculateElapsedTime(p2)
	s, _ := p1.CalculateSpeed(p2)

	if s < 10 {
		return t * idleAmount
	}

	if p1.timestamp.Hour() >= 5 && p2.timestamp.Hour() >= 5 {
		return d * singleAmount
	}

	if p1.timestamp.Hour() < 5 && p2.timestamp.Hour() < 5 {
		return d * increasedAmount
	}

	if p1.timestamp.Hour() >= 5 && p2.timestamp.Hour() < 5 {
		t := time.Date(p2.timestamp.Year(), p2.timestamp.Month(), p2.timestamp.Day(), 0, 0, 0, 0, time.UTC)
		pt1 := math.Abs(p1.timestamp.Sub(t).Hours())
		pt2 := math.Abs(p2.timestamp.Sub(t).Hours())
		return (s * pt1 * singleAmount) + (s * pt2 * increasedAmount)
	}

	return 0.0
}
