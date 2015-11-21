package metrik

type metricListResponse struct {
	Metrics []*Metric `json:"metrics"`
}

type tagListResponse struct {
	Tags []*Tag `json:"tags"`
}

//TotalAggregateResponseItem represents the response that the HTTP/JSON API will send to total aggregate
//queries, eg. /sum/metric. It can be modified using the TotalAggregateHook() method of the Server.
type TotalAggregateResponseItem struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

//GroupbyAggregateResponseItem represents the response that the HTTP/JSON API will send to group-by aggregate
//queries, eg. /sum/metric/by/tag. It can be modified using the GroupbyAggregateHook() method of the Server.
type GroupbyAggregateResponseItem struct {
	Name   string  `json:"name"`
	Groups []group `json:"groups"`
}

//TotalAggregateResponse is an array of aggregated metric responses.
type TotalAggregateResponse struct {
	Metrics []TotalAggregateResponseItem `json:"metrics"`
}

//GroupbyAggregateResponse is an array of group-by aggregated metric responses.
type GroupbyAggregateResponse struct {
	Metrics []GroupbyAggregateResponseItem `json:"metrics"`
}

//TotalAggregateHook is an interface to hook and transform total aggregate responses.
//the return type should be marshallable to JSON.
type TotalAggregateHook func(TotalAggregateResponse) interface{}

//GroupbyAggregateHook is an interface to hook and transform groupby aggregate responses.
//the return type should be marshallable to JSON.
type GroupbyAggregateHook func(GroupbyAggregateResponse) interface{}
