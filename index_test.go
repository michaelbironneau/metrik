package metrik

import (
	"strconv"
	"testing"
)

func dummyIndex1() *invertedIndex {
	m := make(Points, 10000)
	for i := range m {
		m[i] = Point{
			Tags:  map[string][]string{"rack": []string{strconv.Itoa(i % 20)}},
			Value: 1.0,
		}
	}
	i := newInvertedIndex()
	i.Index(m)
	return &i
}

func TestTotal(t *testing.T) {
	index := dummyIndex1()
	val, _ := index.GetTotalAggregate(&sum{}, nil)
	if val != 10000 {
		t.Errorf("expected sum to be 10000, instead got %v", val)
	}
	val, _ = index.GetTotalAggregate(&avg{}, nil)
	if val != 1.0 {
		t.Errorf("expected avg to be 1, instead got %v", val)
	}
	val, _ = index.GetTotalAggregate(&count{}, nil)
	if val != 10000 {
		t.Errorf("expected count to be 10000, instead got %v", val)
	}
}

func TestTotalFiltered(t *testing.T) {
	index := dummyIndex1()
	val, _ := index.GetTotalAggregate(&count{}, map[string][]string{"rack": []string{"0"}})
	if val != 500 {
		t.Errorf("expected count to be 500, instead got %v", val)
	}
	val, _ = index.GetTotalAggregate(&count{}, map[string][]string{"rack": []string{"0", "1"}})
	if val != 0 {
		t.Errorf("expected count to be 0, instead got %v", val)
	}
}

func BenchmarkTotal(b *testing.B) {
	b.StopTimer()
	index := dummyIndex1()
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		index.GetTotalAggregate(&avg{}, nil)
	}
}

func BenchmarkTotalFiltered(b *testing.B) {
	b.StopTimer()
	index := dummyIndex1()
	filter := map[string][]string{"rack": []string{"0"}}
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		index.GetTotalAggregate(&avg{}, filter)
	}
}

func BenchmarkGroupBy(b *testing.B) {
	b.StopTimer()
	index := dummyIndex1()
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		index.GetGroupByAggregate("rack", &avg{}, nil)
	}
}

func BenchmarkGroupByFiltered(b *testing.B) {
	b.StopTimer()
	index := dummyIndex1()
	filter := map[string][]string{"rack": []string{"0"}}
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		index.GetGroupByAggregate("rack", &avg{}, filter)
	}
}
