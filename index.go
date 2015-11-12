package metrik

import (
	"bytes"
	"encoding/gob"
)

type leaf struct {
	Ids  []int //although we're not using ids now, we might use them in future if we introduce a "filter" operator
	Vals []float64
}

//type GroupByGroups represents aggregated metric values, broken down by group-by tags.
type GroupbyGroup struct {
	Key   string  `json:"key"`
	Value float64 `json:"value"`
}

type tagGroup map[string]*leaf

//type invertedIndex is an immutable mapping from tag key-value pairs to arrays of points
//that are tagged with those pairs. Because it's immutable we don't need any locking around it.
type invertedIndex map[string]tagGroup

func newInvertedIndex() invertedIndex {
	ii := make(invertedIndex)
	return ii
}

func (ii invertedIndex) Index(points MetricValue) {
	for i := range points {
		ii.indexPoint(points[i], i)
	}
}

func (ii invertedIndex) indexPoint(point MetricPoint, id int) {
	for tag, values := range point.Tags {
		if tagMap, ok := ii[tag]; ok {
			for _, val := range values {
				if tagVal, ok2 := tagMap[val]; ok2 {
					//append to existing leaf
					tagVal.Ids = append(tagVal.Ids, id)
					tagVal.Vals = append(tagVal.Vals, point.Value)
				} else {
					//new leaf
					ii[tag][val] = &leaf{
						Ids:  []int{id},
						Vals: []float64{point.Value},
					}
				}
			}
		} else {
			//new tag key
			ii[tag] = make(tagGroup)
			for _, val := range values {
				ii[tag][val] = &leaf{
					Ids:  []int{id},
					Vals: []float64{point.Value},
				}
			}
		}
	}
}

func (ii invertedIndex) GetTagGroup(t Tag) (tagGroup, bool) {
	group, ok := ii[t.Name()]
	return group, ok
}

func (ii invertedIndex) GetGroupByAggregate(tag string, a Aggregator) ([]GroupbyGroup, bool) {
	var (
		tg  tagGroup
		ok  bool
		ret []GroupbyGroup
	)
	if tg, ok = ii[tag]; !ok {
		return nil, false
	}

	ret = make([]GroupbyGroup, 0, len(tg))
	for key, values := range tg {
		ret = append(ret, GroupbyGroup{
			Key:   key,
			Value: a.Apply(values.Vals),
		})
	}
	return ret, true
}

func (ii invertedIndex) getTotalAggregate(a Aggregator) float64 {
	var valList [][]float64
	for _, branch := range ii {
		for _, tagLeaf := range branch {
			valList = append(valList, tagLeaf.Vals) //slices are passed by reference so this doesn't actually create a copy of the data...right?
		}
	}
	return a.ApplyMany(valList)
}

func (ii invertedIndex) Marshal() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(ii)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func unmarshalInvertedIndex(b []byte) (invertedIndex, error) {
	var (
		ii invertedIndex
	)
	dec := gob.NewDecoder(bytes.NewBuffer(b))
	err := dec.Decode(&ii)
	return ii, err
}
