package metrik

//type Aggregator reduces a list of values to a single value.
type Aggregator interface {
	Apply([]float64) float64
}

type Sum struct{}

func (s Sum) Apply(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	var ret float64
	for _, val := range vals {
		ret += val
	}
	return ret
}

type Avg struct{}

func (a Avg) Apply(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	var sum float64
	for _, val := range vals {
		sum += val
	}
	return sum / float64(len(vals))
}
