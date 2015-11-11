package metrik

import (
	"encoding/json"
	"net/http"
	"regexp"
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

type Server struct {
	metrics      []Metric
	tags         []Tag
	store        Store
	auth         AuthProvider
	indexHandler http.Handler
	_metricsMeta []MetricMetadata
	_tagsMeta    []TagMetadata
	_mms         []byte
	_tms         []byte
}

func NewServer() *Server {
	s := Server{
		store: &InMemoryStore{},
		auth:  &OpenAPI{},
	}
	return &s
}

func (s *Server) Metric(m Metric) *Server {
	s.metrics = append(s.metrics, m)
	s._metricsMeta = append(s._metricsMeta, getMetricMetadata(m))
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

func (s *Server) metricsIndexHandler(w http.ResponseWriter, r *http.Request) {
	addHeaders(w)
	w.Write(s._mms)
}

func (s *Server) tagsIndexHandler(w http.ResponseWriter, r *http.Request) {
	addHeaders(w)
	w.Write(s._tms)
}

func addHeaders(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)
}

//handles queries of the form GET /metrics/:metric_1[,:metric_2[,...:metric_n]]
func (s *Server) metricTotalAggregateHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.String()[9:] //strip "/metrics/"
	metrics := strings.Split(url, ",")
	var retval metricQueryResponse
}

//handles queries of the form GET /metrics/:metric_1[,:metric_2[,...:metric_n]]/by/:tag
func (s *Server) metricGroupByAggregateHandler(w http.ResponseWriter, r *http.Request) {

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
	handler.Route("/metrics/.+/by/.+", s.metricGroupByAggregateHandler).Route("/metrics/.+", handler)                     //the order of these two matters

	http.ListenAndServe(":"+strconv.Itoa(port), &handler)
	return nil
}
