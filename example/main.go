package main

import (
	"github.com/michaelbironneau/metrik"
	"math/rand"
	"strconv"
	"time"
)

func runUpdater() (chan metrik.MetricValue, chan bool, error) {
	result := make(chan metrik.MetricValue, 1)
	stop := make(chan bool)
	go func() {
		for {
			select {
			case <-time.After(2 * time.Second):
				m := make(metrik.MetricValue, 10)
				for i := range m {
					m[i] = metrik.MetricPoint{
						Tags:  map[string][]string{"rack": []string{strconv.Itoa(i % 3)}},
						Value: rand.NormFloat64()*0.1 + 0.3,
					}
				}
			case <-stop:
				return
			}

		}
	}()
	return result, stop, nil
}

func main() {
	server := metrik.NewServer()
	server.Metric(&metrik.Metric{Name: "cpu", Description: "1-min averaged CPU usage", StartUpdater: runUpdater, Units: "%"})
	server.Tag(&metrik.Tag{Name: "rack", Description: "Rack number of machine"})
	server.Serve(8080)
}
