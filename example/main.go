package main

import (
	"github.com/michaelbironneau/metrik"
	"math/rand"
	"strconv"
	"time"
)

func updateCPU(result chan metrik.Points, stop chan bool) error {
	for {
		select {
		case <-time.After(2 * time.Second):
			m := make(metrik.Points, 10)
			for i := range m {
				m[i] = metrik.Point{
					Tags:  map[string][]string{"rack": []string{strconv.Itoa(i % 3)}},
					Value: rand.NormFloat64()*0.1 + 0.3,
				}
			}
			result <- m
		case <-stop:
			return nil
		}

	}
}

func main() {
	server := metrik.NewServer()
	server.Metric(&metrik.Metric{Name: "cpu", Description: "1-min averaged CPU usage", UpdateFunc: updateCPU, Units: "%"})
	server.Tag(&metrik.Tag{Name: "rack", Description: "Rack number of machine"})
	server.Serve(8080)
}
