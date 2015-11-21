package metrik

//Aggregator reduces a list of values to a single value.
type Aggregator interface {
	Apply([]float64) float64
	ApplyMany([][]float64) float64
}

type count struct{}

func (c count) Apply(vals []float64) float64 {
	return float64(len(vals))
}

func (c count) ApplyMany(vals [][]float64) float64 {
	var ret float64
	for _, list := range vals {
		ret += float64(len(list))
	}
	return ret
}

type sum struct{}

func (s sum) Apply(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	var ret float64
	for _, val := range vals {
		ret += val
	}
	return ret
}

func (s sum) ApplyMany(valLists [][]float64) float64 {
	var ret float64
	for _, vals := range valLists {
		ret += s.Apply(vals)
	}
	return ret
}

type avg struct{}

func (a avg) Apply(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	var ssum float64
	for _, val := range vals {
		ssum += val
	}
	return ssum / float64(len(vals))
}

func (a avg) ApplyMany(valLists [][]float64) float64 {
	var (
		ssum  float64
		count float64
	)
	s := sum{}
	for _, vals := range valLists {
		ssum += s.Apply(vals)
		count += float64(len(vals))
	}
	return ssum / count
}
