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
	Name() string               //Short name for metric. Should be URL-friendly.
	Units() string              //Units for the metric, for example "Kw".
	Description() string        //Description of the metric, for users.
	Pull() (MetricValue, error) //Fetch the latest metric points
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

//type GroupByGroups represents aggregated metric values, broken down by group-by tags.
type GroupbyGroups struct {
	Key   string  `json:"key"`
	Value float64 `json:"value"`
}

//Group the metric points by tag key/value pairs and aggregate with the given aggregator.
func (mv MetricValue) GroupBy(t Tag, a Aggregator) []GroupbyGroups {
	tagValues := t.Enumerate()
	ret := make([]GroupbyGroups, len(tagValues), len(tagValues))
	for i, tagValue := range tagValues {
		ret[i].Key = tagValue
		ret[i].Value = a.Apply(mv.taggedWith(t.Name(), tagValue))
	}
	return ret
}

//Aggregate values with the given aggregator
func (mv MetricValue) Aggregate(a Aggregator) float64 {
	vals := make([]float64, len(mv), len(mv))
	for i, point := range mv {
		vals[i] = point.Value
	}
	return a.Apply(vals)
}

func isIn(find string, collection []string) bool {
	for _, val := range collection {
		if find == val {
			return true
		}
	}
	return false
}

func (mv MetricValue) taggedWith(tagKey string, tagValue string) []float64 {
	var ret []float64
	for _, point := range mv {
		if vals, ok := point.Tags[tagKey]; ok && isIn(tagValue, vals) {
			ret = append(ret, point.Value)
		}
	}
	return ret
}
