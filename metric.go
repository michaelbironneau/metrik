package metrik

//type MetricPoint represents a tagged real-time metric value (e.g. Most recent CPU usage tagged with {"machine": "testserver"})
type MetricPoint struct {
	Tags  Tags
	Value float64
}

//type MetricValue is a collection of metric points over the whole population.
type MetricValue []MetricPoint

//Metric provides an interface between the data fetcher and the aggregator.
type Metric interface {
	Name() string                                     //Short name for metric. Should be URL-friendly.
	Units() string                                    //Units for the metric, for example "Kw".
	Description() string                              //Description of the metric, for users.
	RunUpdater() (chan MetricValue, chan bool, error) //start the updater. It should publish values through the first channel, and accept a stop command on the second.
}

type MetricMetadata struct {
	Name        string `json:"name"`
	Units       string `json:"units"`
	Description string `json:"description"`
}

func getMetricMetadata(m Metric) MetricMetadata {
	return MetricMetadata{
		Name:        m.Name(),
		Units:       m.Units(),
		Description: m.Description(),
	}
}
