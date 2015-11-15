package metrik

type metricListResponse struct {
	Metrics []*Metric `json:"metrics"`
}

type tagListResponse struct {
	Tags []*Tag `json:"tags"`
}

//TotalAggregateItem
type TotalAggregateResponseItem struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

//GroupbyAggregateItem
type GroupbyAggregateResponseItem struct {
	Name   string  `json:"name"`
	Groups []group `json:"groups"`
}

//TotalAggregateResponse
type TotalAggregateResponse struct {
	Metrics []TotalAggregateResponseItem `json:"metrics"`
}

//GroupbyAggregateResponse
type GroupbyAggregateResponse struct {
	Metrics []GroupbyAggregateResponseItem `json:"metrics"`
}

//type TotalAggregateHook is an interface to hook and transform total aggregate responses.
//the return type should be marshallable to JSON.
type TotalAggregateHook func(TotalAggregateResponse) interface{}

//type GroupbyAggregateHook is an interface to hook and transform groupby aggregate responses.
//the return type should be marshallable to JSON.
type GroupbyAggregateHook func(GroupbyAggregateResponse) interface{}
