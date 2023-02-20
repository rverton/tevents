package tevents

import (
	"database/sql"
	"embed"
	"fmt"
	"html"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"

	"tailscale.com/client/tailscale"
)

type Server struct {
	addr     string
	server   *http.Server
	db       *sql.DB
	listener net.Listener
	tsClient *tailscale.LocalClient

	EventService EventService
}

type TplData struct {
	Events      []*Event
	EventGroups map[string][]bool
	Monitor     bool
	LastHours   int
}

//go:embed assets/*
var assetsFS embed.FS

const monitorHours = 48

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

	indexTmpl = parseTpl(funcMap, "assets/events.html")
	indexMonitorTmpl = parseTpl(funcMap, "assets/monitors.html")
}

func NewServer(addr string, db *sql.DB, ln net.Listener, lc *tailscale.LocalClient) *Server {
	return &Server{
		server: &http.Server{
			Addr: addr,
		},
		db:       db,
		listener: ln,
		tsClient: lc,
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
	mux.HandleFunc("/.clear", s.handleClear)

	mux.Handle("/assets/", http.FileServer(http.FS(assetsFS)))

	return mux
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	events, err := s.EventService.Find("event")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	indexTmpl.ExecuteTemplate(w, "layout.html", TplData{Events: events, Monitor: false})
}

func (s *Server) handleIndexMonitor(w http.ResponseWriter, r *http.Request) {
	events, err := s.EventService.Find("monitor")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// group entries by origin to show monitor hooks over time
	eventsGrouped := make(map[string][]*Event)
	for _, e := range events {
		identifier := fmt.Sprintf("%s:%s", e.Origin, e.Owner)
		eventsGrouped[identifier] = append(eventsGrouped[identifier], e)
	}

	// calculate hours for each event group
	eventsGroupedHours := make(map[string][]bool)
	for k, v := range eventsGrouped {
		eventsGroupedHours[k] = MonitorMap(time.Now(), v, monitorHours)
	}

	indexMonitorTmpl.ExecuteTemplate(w, "layout.html", TplData{
		EventGroups: eventsGroupedHours,
		Monitor:     true,
		LastHours:   monitorHours,
	})
}

// MonitorMap groups events by hour in reverse order
func MonitorMap(now time.Time, events []*Event, lastHours int) []bool {

	// hours represents the last hours and if a
	// monitoring event occured in this hour
	hours := make([]bool, lastHours)

	for _, e := range events {
		diff := int(now.Sub(e.CreatedAt).Minutes() / 60)

		if diff > lastHours {
			continue
		}

		hours[diff] = true
	}

	// reverse slice
	for i, j := 0, len(hours)-1; i < j; i, j = i+1, j-1 {
		hours[i], hours[j] = hours[j], hours[i]
	}

	return hours
}

func (s *Server) handleLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	who, err := s.tsClient.WhoIs(r.Context(), r.RemoteAddr)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	owner := html.EscapeString(who.Node.ComputedName)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	origin := r.URL.Query().Get("origin")

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
	who, err := s.tsClient.WhoIs(r.Context(), r.RemoteAddr)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	owner := html.EscapeString(who.Node.ComputedName)

	if err := s.EventService.Insert(origin, "monitor", "", owner); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := s.EventService.ClearAll(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) Start() error {

	s.server.Handler = s.routes()

	log.Println("Starting server on", s.server.Addr)
	return s.server.Serve(s.listener)
}

func truncateToHour(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
}

func parseTpl(funcs template.FuncMap, file string) *template.Template {
	return template.Must(
		template.New("layout.html").Funcs(funcs).ParseFS(assetsFS, "assets/layout.html", file))
}
