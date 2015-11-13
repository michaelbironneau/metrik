package metrik

import "time"

//type MetricPoint represents a tagged real-time metric value (e.g. Most recent CPU usage tagged with {"machine": "testserver"})
type Point struct {
	Tags  Tags
	Value float64
}

//type Points is a collection of metric points over the whole population.
type Points []Point

//Function that runs the updater. It should block. It should publish values through the first channel, and accept a stop command on the second.
type Updater func(chan Points, chan bool) error

//Metric provides an interface between the data fetcher and the aggregator.
type Metric struct {
	Name        string  `json:"name"`        //Short name for metric. Should be URL-friendly.
	Units       string  `json:"units"`       //Units for the metric, for example "Kw".
	Description string  `json:"description"` //Description of the metric, for users.
	UpdateFunc  Updater `json:"-"`
}

//Utility function to convert a periodic polling updater to Updater type
func PollUpdater(fetch func() (Points, error), interval time.Duration) Updater {
	return func(result chan Points, stop chan bool) error {
		for {
			select {
			case <-time.After(interval):
				points, err := fetch()
				if err != nil {
					return err
				}
				result <- points
			case <-stop:
				return nil
			}

		}
	}
}
