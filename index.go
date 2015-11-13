package metrik

import (
	"bytes"
	"encoding/gob"
)

type leaf struct {
	Ids  []int //although we're not using ids now, we might use them in future if we introduce a "filter" operator
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

func (ii invertedIndex) GetGroupByAggregate(tag string, a Aggregator) ([]group, bool) {
	var (
		tg  tagGroup
		ok  bool
		ret []group
	)
	if tg, ok = ii[tag]; !ok {
		return nil, false
	}

	ret = make([]group, 0, len(tg))
	for key, values := range tg {
		ret = append(ret, group{
			Key:   key,
			Value: a.Apply(values.Vals),
		})
	}
	return ret, true
}

//get total aggregate, optionally filtered by tags. the bool return functions as 'ok',
//as in 'ok, we found the tags in the filter'
func (ii invertedIndex) getTotalAggregate(a Aggregator, t Tags) (float64, bool) {
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

	//filter
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
			return 0, false
		}
	}
	return a.Apply(intersection.Vals), true

}

//intersect the lists s1 and s2.
//each point we index has an autoincreasing id. therefore, the intersection
//algorithm is entitled to assume that the lists s1 and s2 are sorted in
//increasing order.
func intersect(l1, l2 leaf) *leaf {
	var ret leaf
	var capacity int
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
