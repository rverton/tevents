package tevents

import (
	"database/sql"
	"embed"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type Server struct {
	addr   string
	server *http.Server
	db     *sql.DB

	EventService EventService
}

//go:embed assets/*
var assetsFS embed.FS

var (
	indexTmpl        *template.Template
	indexMonitorTmpl *template.Template
)

func init() {
	funcMap := template.FuncMap{
		"formatTime": func(t time.Time) string {
			return t.String()[:19]
		},
	}

	indexTmpl = template.Must(template.New("events").Funcs(funcMap).ParseFS(assetsFS, "assets/layout.html", "assets/events.html"))
	indexMonitorTmpl = template.Must(template.New("monitors").Funcs(funcMap).ParseFS(assetsFS, "assets/layout.html", "assets/monitors.html"))
}

func NewServer(addr string, db *sql.DB) *Server {
	return &Server{
		server: &http.Server{
			Addr: addr,
		},
		db: db,
	}
}

func (s *Server) routes() *http.ServeMux {
	mux := http.NewServeMux()

	// display
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/monitor", s.handleIndexMonitor)

	// collect
	mux.HandleFunc("/.log", s.handleLog)
	mux.HandleFunc("/.monitor", s.handleMonitor)

	mux.Handle("/assets/", http.FileServer(http.FS(assetsFS)))

	return mux
}

type TplData struct {
	Events      []*Event
	EventGroups map[string][]*Event
	Monitor     bool
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	events, err := s.EventService.Find("event")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	indexTmpl.ExecuteTemplate(w, "layout", TplData{Events: events, Monitor: false})
}

func (s *Server) handleIndexMonitor(w http.ResponseWriter, r *http.Request) {
	events, err := s.EventService.Find("monitor")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// group entries by origin to show monitor hooks over time
	groups := make(map[string][]*Event)
	for _, e := range events {
		identifier := fmt.Sprintf("%s:%s", e.Origin, e.Owner)
		groups[identifier] = append(groups[identifier], e)
	}

	indexMonitorTmpl.ExecuteTemplate(w, "layout", TplData{
		EventGroups: groups,
		Monitor:     true,
	})
}

func (s *Server) handleLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	origin := r.URL.Query().Get("origin")
	owner := "ts-owner"

	if err := s.EventService.Insert(origin, "event", string(body), owner); err != nil {
		http.Error(w, fmt.Sprintf("sqlite error: %v", err), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleMonitor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	origin := r.URL.Query().Get("origin")
	owner := "ts-owner"

	if err := s.EventService.Insert(origin, "monitor", "", owner); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) Start() error {

	s.server.Handler = s.routes()

	log.Println("Starting server on", s.server.Addr)
	return s.server.ListenAndServe()
}
