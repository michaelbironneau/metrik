package metrik

type metricListResponse struct {
	Metrics []*Metric `json:"metrics"`
}

type tagListResponse struct {
	Tags []*Tag `json:"tags"`
}

type metricQueryResponseItem struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

type metricGroupByResponseItem struct {
	Name   string         `json:"name"`
	Groups []GroupbyGroup `json:"groups"`
}

type metricQueryResponse struct {
	Metrics []metricQueryResponseItem `json:"metrics"`
}

type metricGroupByResponse struct {
	Metrics []metricGroupByResponseItem `json:"metrics"`
}
