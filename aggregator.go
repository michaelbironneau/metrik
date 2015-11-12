package metrik

//type Aggregator reduces a list of values to a single value.
type Aggregator interface {
	Apply([]float64) float64
	ApplyMany([][]float64) float64
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

func (s Sum) ApplyMany(valLists [][]float64) float64 {
	var ret float64
	for _, vals := range valLists {
		ret += s.Apply(vals)
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

func (a Avg) ApplyMany(valLists [][]float64) float64 {
	var (
		sum   float64
		count float64
	)
	s := Sum{}
	for _, vals := range valLists {
		sum += s.Apply(vals)
		count += float64(len(vals))
	}
	return sum / count
}
