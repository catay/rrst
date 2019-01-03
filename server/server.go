package server

import (
	"fmt"
	"github.com/catay/rrst/repository"
	"github.com/fsnotify/fsnotify"
	"log"
	"net"
	"net/http"
)

// Server model
type Server struct {
	http.Server
	RepoHandleStateTrackers []*RepoHandleStateTracker
	totalActiveRequests     int64
	watcher                 *fsnotify.Watcher
}

// RepoHandleStateTracker tracks Tag handle state for the repositories
type RepoHandleStateTracker struct {
	*repository.Repository
	TagHandleStateTrackers map[string]*TagHandleStateTracker
}

// TagHandleState tracks HTTP handler state
//
// Mainly done because we cannot deregister a HTTP handler or
// redefine it.
type TagHandleStateTracker struct {
	Present    bool
	Registered bool
}

func NewRepoHandleStateTracker(repo *repository.Repository) *RepoHandleStateTracker {
	rh := &RepoHandleStateTracker{
		Repository:             repo,
		TagHandleStateTrackers: make(map[string]*TagHandleStateTracker),
	}
	return rh
}

func (rh *RepoHandleStateTracker) ResetPresentFlag() {
	for k, _ := range rh.TagHandleStateTrackers {
		rh.TagHandleStateTrackers[k].Present = false
	}
}

func (rh *RepoHandleStateTracker) UpdatePresentFlag() {
	for _, t := range rh.Tags {
		_, ok := rh.TagHandleStateTrackers[t.Name]
		if ok {
			rh.TagHandleStateTrackers[t.Name].Present = true
		} else {
			rh.TagHandleStateTrackers[t.Name] = &TagHandleStateTracker{true, false}
		}
	}
}

func (rh *RepoHandleStateTracker) refreshHandlers() {
	for k, v := range rh.TagHandleStateTrackers {

		if v.Present && v.Registered {
			log.Printf("Tag %v already registered ", k)
		}

		if v.Present && !v.Registered {
			// register handle to serve the metadata
			serveMdPath := "/" + rh.ContentSuffixPath + "/" + k + "/repodata/"
			localMdPath := rh.ContentTagsPath + "/" + k + "/repodata/"

			http.Handle(serveMdPath, HTTPLogger(
				rh.serveTag(http.StripPrefix(serveMdPath,
					http.FileServer(http.Dir(localMdPath))), k),
			))

			// register handle to serve the files
			serveFilesPath := "/" + rh.ContentSuffixPath + "/" + k + "/"
			localFilesPath := rh.ContentFilesPath + "/"

			http.Handle(serveFilesPath, HTTPLogger(
				rh.serveTag(http.StripPrefix(serveFilesPath,
					http.FileServer(http.Dir(localFilesPath))), k),
			))

			rh.TagHandleStateTrackers[k].Registered = true
			log.Println("register " + k + " url: " + serveFilesPath)
		}
	}
}

func (rh *RepoHandleStateTracker) serveTag(h http.Handler, tag string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if rh.TagHandleStateTrackers[tag].Present {
			h.ServeHTTP(w, r)
		} else {
			http.NotFound(w, r)
		}
	})
}

// NewServer returns a new server.
func NewServer(port string, repositories []*repository.Repository) *Server {
	s := &Server{
		Server: http.Server{
			Addr: ":" + port,
		},
	}

	// FIXME: make some telemetry struct to track this
	s.ConnState = func(n net.Conn, h http.ConnState) {
		if h == http.StateActive {
			s.totalActiveRequests++
		}
	}

	// init repo handle state tracker
	for i, _ := range repositories {
		s.RepoHandleStateTrackers = append(s.RepoHandleStateTrackers, NewRepoHandleStateTracker(repositories[i]))
	}

	return s
}

func (s *Server) TagWatcher() {
	for {
		e, _ := <-s.watcher.Events
		log.Println("fs event" + e.String())

		s.registerHandlers()
	}
}

func (s *Server) config(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Total requests served: %v\n\n", s.totalActiveRequests)
	for i, _ := range s.RepoHandleStateTrackers {

		fmt.Fprintf(w, "* %v\n", s.RepoHandleStateTrackers[i].Name)
		fmt.Fprintf(w, "%-10v %-10v %v\n", "Handle", "Present", "Registered")
		for k, v := range s.RepoHandleStateTrackers[i].TagHandleStateTrackers {
			fmt.Fprintf(w, "%-10v %-10v %v\n", k, v.Present, v.Registered)

		}
		fmt.Fprintf(w, "\n")
	}
}

func (s *Server) Run() error {
	log.Printf("Start server: %v", s.Addr)

	var err error

	// Register all the HTTP handlers for the repositories
	s.registerHandlers()

	s.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	for i, _ := range s.RepoHandleStateTrackers {
		if err := s.watcher.Add(s.RepoHandleStateTrackers[i].ContentTagsPath); err != nil {
			log.Println(err, s.RepoHandleStateTrackers[i].ContentTagsPath)
		}
	}

	go s.TagWatcher()

	http.HandleFunc("/config", s.config)

	return s.ListenAndServe()
}

func (s *Server) resetPresentFlag() {
	for i, _ := range s.RepoHandleStateTrackers {
		s.RepoHandleStateTrackers[i].ResetPresentFlag()
	}
}

func (s *Server) updatePresentFlag() {
	for i, _ := range s.RepoHandleStateTrackers {
		s.RepoHandleStateTrackers[i].UpdatePresentFlag()
	}
}

func (s *Server) refreshHandlers() {
	for i, _ := range s.RepoHandleStateTrackers {
		s.RepoHandleStateTrackers[i].refreshHandlers()
	}
}

// registerHandlers registers all the handlers.
func (s *Server) registerHandlers() {
	// Set Present boolean in all the tag handles to false
	s.resetPresentFlag()

	// refresh the repository state
	for i, _ := range s.RepoHandleStateTrackers {
		s.RepoHandleStateTrackers[i].RefreshState()
	}

	// Check which files are in the map and put presence to true.
	s.updatePresentFlag()

	s.refreshHandlers()
}

// HTTPLogger is a custom HTTP request logger.
// It takes a http handler as an argument, outputs  and returns it unmodified.
func HTTPLogger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s %s %s\n", r.RemoteAddr, r.Method, r.URL, r.Proto, r.UserAgent())
		handler.ServeHTTP(w, r)
	})
}
