package metrik

import (
	"bytes"
	"encoding/gob"
	"sort"
	"time"
)

type leaf struct {
	Ids  []int
	Vals []float64
}

//type groups represents aggregated metric values, broken down by group-by tags.
type group struct {
	Key   string  `json:"key"`
	Value float64 `json:"value"`
}

type tagGroup map[string]*leaf

//type invertedIndex is an immutable mapping from tag key-value pairs to arrays of points
//that are tagged with those pairs. Because it's immutable we don't need any locking around it.
type invertedIndex map[string]tagGroup

type timeSeriesItem struct {
	T     time.Time
	Index invertedIndex
}

type timeSeries []*timeSeriesItem

func (t timeSeries) Latest() *timeSeriesItem {

	if len(t) == 0 {
		return nil
	}

	return t[len(t)-1]
}

func (t timeSeries) GreaterThan(start time.Time) []*timeSeriesItem {
	i := sort.Search(len(t), func(i int) bool {
		return t[i].T.After(start)
	})
	return t[i:]
}

func (t timeSeries) Insert(timestamp time.Time, index invertedIndex) {
	i := sort.Search(len(t), func(i int) bool {
		return t[i].T.After(timestamp)
	})
	
	switch {
		case i == len(t) && len(t) < cap(t):
			//append
			t = append(t, &timeSeriesItem{timestamp, index})
		case i == len(t) && len(t) == cap(t):
			//truncate and append
			t = append(t[1:], &timeSeriesItem{timestamp, index})
		case i == 0 && len(t) < cap(t):
			//prepend
			t = append([]*timeSeriesItem{&timeSeriesItem{timestamp, index}}, t...)
		case i == 0 && len(t) == cap(t):
			//can't prepend since the buffer is full 
			//ignore the request
			return
		default:
			//somewhere in the middle
			if len(t) < cap(t) {
				t = append(t[:i], append([]*timeSeriesItem{&timeSeriesItem{timestamp, index}}, t[i+1:]...)...)
			} else {
				//truncate and insert
				t = append(t[1:i], append([]*timeSeriesItem{&timeSeriesItem{timestamp, index}}, t[i+1:]...)...)
			}
	}		
}

//fast path in case the timestamp is left to wall time
func (t timeSeries) Append(index invertedIndex) {
	t = append(t, &timeSeriesItem{time.Now(), index})
}

func newInvertedIndex() invertedIndex {
	ii := make(invertedIndex)
	return ii
}

func (ii invertedIndex) Index(points Points) {
	for i := range points {
		ii.indexPoint(points[i], i)
	}
}

func (ii invertedIndex) indexPoint(point Point, id int) {
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
	group, ok := ii[t.Name]
	return group, ok
}

func (ii invertedIndex) GetGroupByAggregate(tag string, a Aggregator, t Tags) ([]group, bool) {
	var (
		tg             tagGroup
		ok             bool
		ret            []group
		filter         *leaf
		filteredValues *leaf
	)
	if tg, ok = ii[tag]; !ok {
		return nil, false
	}
	if t != nil && len(t) > 0 {
		filter, ok = ii.filter(t)
		if !ok {
			return nil, false
		}

	}

	ret = make([]group, 0, len(tg))
	for key, values := range tg {
		if filterVals, ok := t[tag]; ok && !isIn(filterVals, key) {
			//short-circuit if we are filtering by the group-by
			//tag and the filter value doesn't match the current
			//leaf tag value. Eg. group by rack where rack = 0
			//only matches one leaf.
			ret = append(ret, group{
				Key:   key,
				Value: 0,
			})
			continue
		}
		if filter == nil {
			filteredValues = values
		} else {
			filteredValues = intersect(*filter, *values)
		}
		ret = append(ret, group{
			Key:   key,
			Value: a.Apply(filteredValues.Vals),
		})
	}
	return ret, true
}

func isIn(slice []string, search string) bool {
	for _, s := range slice {
		if s == search {
			return true
		}
	}
	return false
}

//get total aggregate, optionally filtered by tags. the bool return functions as 'ok',
//as in 'ok, we found the tags in the filter'
func (ii invertedIndex) GetTotalAggregate(a Aggregator, t Tags) (float64, bool) {
	var valList [][]float64

	//no filter
	if t == nil || len(t) == 0 {
		for _, branch := range ii {
			for _, tagLeaf := range branch {
				valList = append(valList, tagLeaf.Vals) //slices are passed by reference so this doesn't actually create a copy of the data...right?
			}
		}
		return a.ApplyMany(valList), true
	}

	if filtered, ok := ii.filter(t); ok {
		return a.Apply(filtered.Vals), true
	}

	return 0, false

}

func (ii invertedIndex) filter(t Tags) (*leaf, bool) {
	var intersection *leaf
	for tagKey, tagValues := range t {
		if leaves, ok := ii[tagKey]; ok {
			for _, val := range tagValues {
				if l, ok2 := leaves[val]; ok2 && intersection == nil {
					//initialize intersection
					intersection = l
				} else if ok2 && intersection != nil {
					intersection = intersect(*intersection, *l)
				}
			}
		} else {
			return nil, false
		}
	}
	return intersection, true
}

//intersect the lists s1 and s2.
//each point we index has an autoincreasing id. therefore, the intersection
//algorithm is entitled to assume that the lists s1 and s2 are sorted in
//increasing order.
func intersect(l1, l2 leaf) *leaf {
	var ret leaf
	var capacity int
	if len(l1.Ids) == 0 || len(l2.Ids) == 0 {
		return &ret
	}
	//set capacity to be the smaller of the two lists
	if len(l1.Ids) > len(l2.Ids) {
		capacity = len(l2.Ids)
	} else {
		capacity = len(l1.Ids)
	}
	ret.Ids = make([]int, 0, capacity)
	ret.Vals = make([]float64, 0, capacity)
	for i, s := range l1.Ids {
		for _, t := range l2.Ids {
			if t > s {
				break
			}
			if s == t {
				ret.Ids = append(ret.Ids, s)
				ret.Vals = append(ret.Vals, l1.Vals[i]) //doesn't matter if we use l1.Vals or l2.Vals
				break
			}
		}
	}
	return &ret
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
