package metrik

//type MetricPoint represents a tagged real-time metric value (e.g. Most recent CPU usage tagged with {"machine": "testserver"})
type MetricPoint struct {
	Tags  Tags
	Value float64
}

//type MetricValue is a collection of metric points over the whole population.
type MetricValue []MetricPoint

//Metric provides an interface between the data fetcher and the aggregator.
type Metric struct {
	Name        string                                  `json:"name"`        //Short name for metric. Should be URL-friendly.
	Units       string                                  `json:"units"`       //Units for the metric, for example "Kw".
	Description string                                  `json:"description"` //Description of the metric, for users.
	UpdateFunc  func(chan MetricValue, chan bool) error `json:"-"`           //Function that runs the updater. It should block. It should publish values through the first channel, and accept a stop command on the second.
}
