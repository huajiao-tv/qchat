package stat

import (
	"net/http"
	"sync/atomic"
	"encoding/json"
	"sync"
)

const (
	LVSCheck = "/status.php"
)

type HttpResponser interface {
	Response() []byte
}

type ErrorResponse struct {
	Code   int         `json:"code"`
	Reason string      `json:"reason"`
	Data   interface{} `json:"data,omitempty"`
}

func (r *ErrorResponse) Response() []byte {
	s, e := json.Marshal(r)
	if e != nil {
		return []byte(`{"code":500,"reason":"Marshal failed ` + e.Error() + `","data":""}`)
	}
	return s
}

type Handler func(*http.Request) HttpResponser

type HttpStat struct {
	*Stat
	work int32

	l sync.RWMutex
	handlers map[string]Handler
}

func NewHttpStat(interval int) *HttpStat {
	return &HttpStat{
		Stat: NewStat(interval),
		work: 0,
	}
}

func (s *HttpStat) IsWorking() bool {
	return atomic.LoadInt32(&s.work) == 0
}

func (s *HttpStat) Up() {
	atomic.StoreInt32(&s.work, 0)
}

func (s *HttpStat) Down() {
	atomic.StoreInt32(&s.work, 1)
}

func (s *HttpStat) HandleFunc(pattern string, handler Handler) {
	s.l.Lock()
	s.handlers[pattern] = handler
	s.l.Unlock()
}

func (s *HttpStat) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// @todo: 白名单

	//start := time.Now()
	var resp HttpResponser
	pathInfo := r.URL.Path
	s.Incr(pathInfo)
	switch pathInfo {
	case LVSCheck:
	default:
		s.l.RLock()
		resp = s.handlers[pathInfo](r)
		s.l.Unlock()
	}
	// @todo: add access log
	w.Write(resp.Response())
}
