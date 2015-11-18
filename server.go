package metrik

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const _DEFAULT_INDEX = `
<!doctype html>
<html>
<head>
	<title>Metrik server default index page</title>
</head>
<body>
	<p>If you're seeing this, a Metrik server is running. Replace this by your own index page using Server.DefaultIndex(yourHandler).</p>
	<p>To learn more about Metrik, please visit <a href="https://www.github.com/michaelbironneau/metrik">the Github repo</a></p>
</body>
</html>
`

const _UNKNOWN_AGGREGATE = `
{"error": "unknown aggregate"}
`

const _INTERNAL_ERROR = `
{"error": "internal server error"}
`

const _UNAUTHORIZED = `
{"error": "_UNAUTHORIZED"}
`

const _NOT_FOUND = `
{"error": "unknown route"}
`

type Server struct {
	metrics      []*Metric
	tags         []*Tag
	store        Store
	auth         AuthProvider
	aggregates   map[string]Aggregator
	indexHandler func(http.ResponseWriter, *http.Request)
	logger       *log.Logger
	taHook       TotalAggregateHook
	gbHook       GroupbyAggregateHook
	_tagsMeta    []Tag
	_mms         []byte
	_tms         []byte
	_indexes     map[string]invertedIndex
	_ilocks      map[string]*sync.RWMutex
	_stopChans   []chan bool
	_updateChans map[string]chan Points
}

func NewServer() *Server {
	s := Server{
		store:        &inMemoryStore{},
		auth:         &openAPI{},
		aggregates:   map[string]Aggregator{"sum": sum{}, "average": avg{}, "count": count{}},
		indexHandler: defaultIndexHandler,
	}
	return &s
}

//set a hook to transform total aggregate response
func (s *Server) TotalAggregateHook(f TotalAggregateHook) *Server {
	s.taHook = f
	return s
}

//set a hook to transform group by aggregate hook
func (s *Server) GroupbyAggregateHook(f GroupbyAggregateHook) *Server {
	s.gbHook = f
	return s
}

//set the logger for the Metrik server.
func (s *Server) Logger(l *log.Logger) *Server {
	s.logger = l
	return s
}

//register a metric.
func (s *Server) Metric(m *Metric) *Server {
	s.metrics = append(s.metrics, m)
	s.logf("added metric %s", m.Name)
	return s
}

//register an aggregate.
func (s *Server) Aggregate(a Aggregator, name string) *Server {
	s.aggregates[name] = a
	s._indexes[name] = newInvertedIndex()
	return s
}

//register a tag group
func (s *Server) Tag(t *Tag) *Server {
	s.tags = append(s.tags, t)
	s.logf("added tag %s", t.Name)
	return s
}

//register a store (currently unused).
func (s *Server) Store(st Store) *Server {
	s.store = st
	return s
}

//register an authentication provider.
func (s *Server) Auth(a AuthProvider) *Server {
	s.auth = a
	return s
}

func defaultIndexHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte(_DEFAULT_INDEX))
}

func unknownAggregateHandler(w http.ResponseWriter, r *http.Request) {
	addHeaders(w, 404)
	w.Write([]byte(_UNKNOWN_AGGREGATE))
}

func (s *Server) metricsIndexHandler(w http.ResponseWriter, r *http.Request) {
	addHeaders(w, 200)
	w.Write(s._mms)
}

func (s *Server) tagsIndexHandler(w http.ResponseWriter, r *http.Request) {
	addHeaders(w, 200)
	w.Write(s._tms)
}

func catchallHandler(w http.ResponseWriter, r *http.Request) {
	addHeaders(w, 404)
	w.Write([]byte(_NOT_FOUND))
}

func addHeaders(w http.ResponseWriter, status int) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
}

func (s *Server) findMetric(name string) (*Metric, bool) {
	for _, metric := range s.metrics {
		if strings.ToLower(metric.Name) == name {
			return metric, true
		}
	}
	return nil, false
}

//wrapper to handle total aggregate queries
//handles queries of the form GET /:aggregate/:metric_1[,:metric_2[,...:metric_n]]
func (s *Server) totalAggHandlerWrapper(aggregate string) func(http.ResponseWriter, *http.Request) {
	var (
		agg      Aggregator
		aggFound bool
	)
	if agg, aggFound = s.aggregates[aggregate]; !aggFound {
		//this should never get reached
		panic("url was matched by regexp but clearly does not satisfy it")
	}
	return func(w http.ResponseWriter, r *http.Request) {
		metrics := strings.Split(r.URL.Path[len(aggregate)+2:], ",") // /sum/a,b,c -> [a,b,c]
		if ok, err := s.auth.Authorize(makeAuthRequest(r, metrics, nil)); !ok && err == nil {
			addHeaders(w, 403)
			w.Write([]byte(_UNAUTHORIZED))
			return
		} else if err != nil {
			addHeaders(w, 500)
			w.Write([]byte(_INTERNAL_ERROR))
			return
		}
		var retval TotalAggregateResponse
		retval.Metrics = make([]TotalAggregateResponseItem, 0, len(metrics))
		for _, metricName := range metrics {
			if index, ok := s._indexes[metricName]; ok {
				filter := parseFilter(r.URL)
				s._ilocks[metricName].RLock()
				val, tagsFound := index.GetTotalAggregate(agg, filter)
				s._ilocks[metricName].RUnlock()
				if tagsFound == false {
					addHeaders(w, 404)
					w.Write([]byte("{\"error\": \"one or more tags in predicate not found\"}"))
					return
				}
				retval.Metrics = append(retval.Metrics, TotalAggregateResponseItem{
					Name:  metricName,
					Value: val,
				})

			} else {
				addHeaders(w, 404)
				w.Write([]byte("{\"error\": \"metric not found - " + metricName + "\"}"))
				return
			}
		}
		addHeaders(w, 200)
		if s.taHook != nil {
			newResponse := s.taHook(retval)
			b, err := json.Marshal(newResponse)
			if err != nil {
				addHeaders(w, 500)
				s.logf("error in hooked total aggregate %v", err)
				w.Write([]byte(_INTERNAL_ERROR))
				return
			}
			w.Write(b)
		}
		b, err := json.Marshal(retval)
		if err != nil {
			addHeaders(w, 500)
			s.logf("error in total aggregate %v", err)
			w.Write([]byte(_INTERNAL_ERROR))
			return
		}
		w.Write(b)
	}
}

func parseFilter(u *url.URL) Tags {
	return Tags(u.Query())
}

func makeAuthRequest(r *http.Request, metrics []string, tags []string) *AuthRequest {
	user, pass, _ := r.BasicAuth()
	return &AuthRequest{
		User:     user,
		Password: pass,
		Metrics:  metrics,
		Tags:     tags,
	}
}

//handles queries of the form GET /metrics/:metric_1[,:metric_2[,...:metric_n]]/by/:tag
func (s *Server) metricGroupByHandlerWrapper(aggregate string) func(http.ResponseWriter, *http.Request) {
	var (
		agg      Aggregator
		aggFound bool
	)
	if agg, aggFound = s.aggregates[aggregate]; !aggFound {
		//this should never get reached
		panic("url was matched by regexp but clearly does not satisfy it")
	}
	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path[len(aggregate)+2:], "/by/")
		if len(parts) != 2 {
			//we should never reach this
			panic("url was matched by regexp but clearly does not satisfy it")
		}
		metricString, tag := parts[0], parts[1]
		metrics := strings.Split(metricString, ",")
		if ok, err := s.auth.Authorize(makeAuthRequest(r, metrics, []string{tag})); !ok && err == nil {
			addHeaders(w, 403)
			w.Write([]byte(_UNAUTHORIZED))
			return
		} else if err != nil {
			addHeaders(w, 500)
			w.Write([]byte(_INTERNAL_ERROR))
			return
		}
		var retval GroupbyAggregateResponse
		retval.Metrics = make([]GroupbyAggregateResponseItem, 0, len(metrics))
		for _, metricName := range metrics {
			if index, ok := s._indexes[metricName]; ok {
				filter := parseFilter(r.URL)
				s._ilocks[metricName].RLock()
				groups, tagFound := index.GetGroupByAggregate(tag, agg, filter)
				s._ilocks[metricName].RUnlock()
				if !tagFound {
					addHeaders(w, 404)
					w.Write([]byte("{\"error\": \"tag not found - " + tag + "\"}"))
					return
				}
				retval.Metrics = append(retval.Metrics, GroupbyAggregateResponseItem{
					Name:   metricName,
					Groups: groups,
				})
			} else {
				addHeaders(w, 404)
				w.Write([]byte("{\"error\": \"metric not found - " + metricName + "\"}"))
				return
			}
		}
		if s.gbHook != nil {
			newResponse := s.gbHook(retval)
			addHeaders(w, 200)
			b, err := json.Marshal(newResponse)
			if err != nil {
				addHeaders(w, 500)
				s.logf("error in hooked groupby aggregate %v", err)
				w.Write([]byte(_INTERNAL_ERROR))
				return
			}
			w.Write(b)
		}
		addHeaders(w, 200)
		b, err := json.Marshal(retval)
		if err != nil {
			addHeaders(w, 500)
			s.logf("error in groupby aggregate %v", err)
			w.Write([]byte(_INTERNAL_ERROR))
			return
		}
		w.Write(b)
	}
}

func (s *Server) logf(fmt string, vals ...interface{}) {
	if s.logger != nil {
		s.logger.Printf(fmt, vals)
	}
}

func (s *Server) updaterWrapper(m *Metric) (chan Points, chan bool) {
	result := make(chan Points, 1)
	stop := make(chan bool)
	go func() {
		for {
			err := m.UpdateFunc(result, stop)
			if err != nil {
				s.logf("updater %s exited with error: %v. retrying in 3 seconds... \n", m.Name, err)
				time.Sleep(3 * time.Second) //Todo: Add some better retry logic here
			} else {
				break
			}
		}
		s.logf("updater %s exited", m.Name)
	}()
	return result, stop
}

//Start metric updaters.
func (s *Server) startUpdaters() error {
	s._updateChans = make(map[string]chan Points)
	s._stopChans = make([]chan bool, 0, len(s.metrics))
	s._ilocks = make(map[string]*sync.RWMutex)
	for _, metric := range s.metrics {
		s.logf("starting updater for %s", metric.Name)
		update, stop := s.updaterWrapper(metric)
		s._updateChans[metric.Name] = update
		s._stopChans = append(s._stopChans, stop)
		s._ilocks[metric.Name] = &sync.RWMutex{}
	}
	s._indexes = make(map[string]invertedIndex)

	go s.listenForChanges()
	return nil
}

func (s *Server) listenForChanges() {
	for {
		for metric, ch := range s._updateChans {
			select {
			case newPoints := <-ch:
				s.logf("received update for metric %s", metric)
				s._ilocks[metric].Lock()
				s._indexes[metric] = newInvertedIndex()
				s._indexes[metric].Index(newPoints)
				s._ilocks[metric].Unlock()
			default:
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}

//Sends a stop signal to the metric updaters.
func (s *Server) StopUpdaters() {
	for _, stopChan := range s._stopChans {
		stopChan <- true
	}
}

func (s *Server) Serve(port int) error {
	var err error
	mms := metricListResponse{
		Metrics: s.metrics,
	}
	tts := tagListResponse{
		Tags: s.tags,
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
		handler.Route("/("+aggregateName+")/(.+)/by/(.+)/*", s.metricGroupByHandlerWrapper(aggregateName))
		handler.Route("/("+aggregateName+")/(.+)/*", s.totalAggHandlerWrapper(aggregateName))
		s.logf("Added aggregate %s", aggregateName)
	}

	handler.Route("/.+/.+", unknownAggregateHandler)
	handler.Route("/", catchallHandler)

	if err := s.startUpdaters(); err != nil {
		return err
	}

	http.ListenAndServe(":"+strconv.Itoa(port), &handler)
	return nil
}
