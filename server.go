package metrik

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

const DEFAULT_INDEX = `
<!doctype html>
<html>
<head>
	<title>Metrik server default index page</title>
</head>
<body>
	<p>If you're seeing this, a Metrik server is running. Replace this by your own index page using Server.DefaultIndex(yourHandler).</p>
</body>
</html>
`

const UNKNOWN_AGGREGATE = `
{"error": "unknown aggregate"}
`

const INTERNAL_ERROR = `
{"error": "internal server error"}
`

type Server struct {
	metrics      []Metric
	tags         []Tag
	store        Store
	auth         AuthProvider
	aggregates   map[string]Aggregator
	indexHandler func(http.ResponseWriter, *http.Request)
	_metricsMeta []MetricMetadata
	_tagsMeta    []TagMetadata
	_mms         []byte
	_tms         []byte
}

func NewServer() *Server {
	s := Server{
		store:      &InMemoryStore{},
		auth:       &OpenAPI{},
		aggregates: map[string]Aggregator{"sum": Sum{}, "avg": Avg{}},
	}
	return &s
}

func (s *Server) Metric(m Metric) *Server {
	s.metrics = append(s.metrics, m)
	s._metricsMeta = append(s._metricsMeta, getMetricMetadata(m))
	return s
}

func (s *Server) Aggregate(a Aggregator, name string) *Server {
	s.aggregates[name] = a
	return s
}

func (s *Server) Tag(t Tag) *Server {
	s.tags = append(s.tags, t)
	s._tagsMeta = append(s._tagsMeta, getTagMetadata(t))
	return s
}

func (s *Server) Store(st Store) *Server {
	s.store = st
	return s
}

func (s *Server) Auth(a AuthProvider) *Server {
	s.auth = a
	return s
}

func defaultIndexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(DEFAULT_INDEX))
}

func unknownAggregateHandler(w http.ResponseWriter, r *http.Request) {
	addHeaders(w, 404)
	w.Write([]byte(UNKNOWN_AGGREGATE))
}

func (s *Server) metricsIndexHandler(w http.ResponseWriter, r *http.Request) {
	addHeaders(w, 200)
	w.Write(s._mms)
}

func (s *Server) tagsIndexHandler(w http.ResponseWriter, r *http.Request) {
	addHeaders(w, 200)
	w.Write(s._tms)
}

func addHeaders(w http.ResponseWriter, status int) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
}

func (s *Server) findMetric(name string) (Metric, bool) {
	for _, metric := range s.metrics {
		if strings.ToLower(metric.Name()) == name {
			return metric, true
		}
	}
	return nil, false
}

//wrapper to handle total aggregate queries
//handles queries of the form GET /:aggregate/:metric_1[,:metric_2[,...:metric_n]]
func (s *Server) totalAggHandlerWrapper(aggregate string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics := strings.Split(r.URL.String()[len(aggregate)+2:], ",") // /sum/a,b,c -> [a,b,c]
		var retval metricQueryResponse
		retval.Metrics = make([]metricQueryResponseItem, len(metrics), len(metrics))
		for _, metricName := range metrics {
			if metric, ok := s.findMetric(metricName); ok {

			} else {
				addHeaders(w, 404)
				w.Write([]byte("{\"error\": \"metric not found - " + metricName + "\"}"))
				return
			}
		}
	}
}

//handles queries of the form GET /metrics/:metric_1[,:metric_2[,...:metric_n]]/by/:tag
func (s *Server) metricGroupByHandlerWrapper(aggregate string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.String()[len(aggregate)+2:], "/by/")
		if len(parts) != 2 {
			//we should never reach this
			panic("url was matched by regexp but clearly does not satisfy it")
		}
		metricString, tag := parts[0], parts[1]
		metrics := strings.Split(metricString, ",")
		var retval metricGroupByResponse
		for _, metricName := range metrics {
			if metric, ok := s.findMetric(metricName); ok {

			} else {
				addHeaders(w, 404)
				w.Write([]byte("{\"error\": \"metric not found - " + metricName + "\"}"))
				return
			}
		}
	}
}

func (s *Server) Serve(port int) error {
	var err error
	mms := metricListResponse{
		Metrics: s._metricsMeta,
	}
	tts := tagListResponse{
		Tags: s._tagsMeta,
	}
	//Cache all the metadata
	if s._mms, err = json.Marshal(mms); err != nil {
		return err
	}
	if s._tms, err = json.Marshal(tts); err != nil {
		return err
	}

	handler := regexpHandler{}
	handler.Route("/$", s.indexHandler).Route("/metrics/*$", s.metricsIndexHandler).Route("/tags/*$", s.tagsIndexHandler) //metadata, the order doesn't matter

	for aggregateName, _ := range s.aggregates {
		handler.Route("/("+aggregateName+")/(.+)/by/(.+)", s.metricGroupByHandlerWrapper(aggregateName))
		handler.Route("/("+aggregateName+")/(.+)", s.totalAggHandlerWrapper(aggregateName))
	}

	handler.Route("/.+/.+", unknownAggregateHandler)
	http.ListenAndServe(":"+strconv.Itoa(port), &handler)
	return nil
}
