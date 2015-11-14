package main

import (
	"github.com/michaelbironneau/metrik"
	"math/rand"
	"strconv"
	"time"
)

func RandomCPU() (metrik.Points, error) {
	m := make(metrik.Points, 10)
	for i := range m {
		m[i] = metrik.Point{
			Tags:  map[string][]string{"rack": []string{strconv.Itoa(i % 3)}},
			Value: rand.NormFloat64()*0.1 + 0.3,
		}
	}
	return m, nil
}

func RandomMemory() (metrik.Points, error) {
	m := make(metrik.Points, 10)
	for i := range m {
		m[i] = metrik.Point{
			Tags:  map[string][]string{"rack": []string{strconv.Itoa(i % 3)}},
			Value: rand.NormFloat64()*100 + 500,
		}
	}
	return m, nil
}

func main() {
	server := metrik.NewServer()
	server.Metric(&metrik.Metric{Name: "cpu", Description: "1-min averaged CPU usage", UpdateFunc: metrik.PollUpdater(RandomCPU, time.Second*3), Units: "%"})
	server.Metric(&metrik.Metric{Name: "memory", Description: "1-min averaged memory usage", UpdateFunc: metrik.PollUpdater(RandomMemory, time.Second*3), Units: "MB"})
	server.Tag(&metrik.Tag{Name: "rack", Description: "Rack number of machine"})
	server.Serve(8080)
}
