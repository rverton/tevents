package tevents

import (
	"database/sql"
	"embed"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
)

type Server struct {
	addr   string
	server *http.Server
	db     *sql.DB

	EventService EventService
}

//go:embed assets/*
var assetsFS embed.FS

var indexTmpl *template.Template

func init() {
	indexTmpl = template.Must(template.ParseFS(assetsFS, "assets/index.html"))
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

	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/log", s.handleLog)
	mux.HandleFunc("/monitor", s.handleMonitor)

	mux.Handle("/assets/", http.FileServer(http.FS(assetsFS)))

	return mux
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	events, err := s.EventService.Find()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	indexTmpl.Execute(w, events)
}

func (s *Server) handleLog(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	origin := r.URL.Query().Get("origin")
	owner := "ts-owner"

	if err := s.EventService.Insert(origin, "log", string(body), owner); err != nil {
		http.Error(w, fmt.Sprintf("sqlite error: %v", err), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleMonitor(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	origin := r.PostForm.Get("origin")
	body := r.PostForm.Get("body")
	owner := r.PostForm.Get("owner")

	if err := s.EventService.Insert(origin, "log", body, owner); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) Start() error {

	s.server.Handler = s.routes()

	log.Println("Starting server on", s.server.Addr)
	return s.server.ListenAndServe()
}
