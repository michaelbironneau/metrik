package metrik

type metricListResponse struct {
	Metrics []MetricMetadata `json:"metrics"`
}

type tagListResponse struct {
	Tags []TagMetadata `json:"tags"`
}

type metricQueryResponseItem struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

type metricGroupByResponseItem struct {
	Name   string          `json:"name"`
	Groups []GroupbyGroups `json:"groups"`
}

type metricQueryResponse struct {
	Metrics []metricQueryResponseItem `json:"metrics"`
}

type metricGroupByResponse struct {
	Metrics []metricGroupByResponseItem `json:"metrics"`
}
