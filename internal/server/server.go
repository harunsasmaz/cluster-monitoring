package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/harunsasmaz/cluster-monitoring/pkg/cluster"
)

type Server struct {
	Addr          string
	MustClose     bool
	server        *http.Server
	clusterClient *cluster.Client
}

func New(client *cluster.Client) *Server {
	port, ok := os.LookupEnv("PORT")
	if !ok {
		log.Fatalf("failed to retrieve port from env")
	}

	server := &Server{
		Addr:          fmt.Sprintf("0.0.0.0:%s", port),
		clusterClient: client,
		MustClose:     true,
	}

	server.initRoutes()
	return server
}

func (s *Server) Serve() error {
	if err := s.server.ListenAndServe(); err != nil {
		if err != http.ErrServerClosed {
			log.Fatalf("error while closing the server: %v\n", err)
		}
	}

	return nil
}

func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.ShutdownWithContext(ctx)
}

func (s *Server) ShutdownWithContext(ctx context.Context) error {
	if err := s.server.Shutdown(ctx); err != nil {
		if s.MustClose {
			s.server.Close()
		}
		return err
	}

	return nil
}

func (s *Server) initRoutes() {
	router := mux.NewRouter()
	router.HandleFunc("/health", s.healthHandler)
	router.HandleFunc("/services/{namespace}", s.corsHandler(s.servicesHandler))
	router.HandleFunc("/services/{namespace}/{group}", s.corsHandler(s.groupHandler))

	s.server = &http.Server{Addr: s.Addr, Handler: router}
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Content-Type", "application/json")
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func (s *Server) corsHandler(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w)
		if r.Method == http.MethodOptions {
			return
		}

		h(w, r)
	}
}

func (s *Server) servicesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	namespace := vars["namespace"]
	services, err := s.clusterClient.GetServicesWithNamespace(context.Background(), namespace)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(HttpError("could not get service information: %v\n", err))
		return
	}

	response, err := json.Marshal(&HttpResponse{
		Data:    services,
		Success: true,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(HttpError("could not marshal response data: %v\n", err))
		return
	}

	w.Write(response)
}

func (s *Server) groupHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	namespace := vars["namespace"]
	groupParam := cluster.LabelParams{
		Label: cluster.ApplicationGroupLabel,
		Value: vars["group"],
	}

	services, err := s.clusterClient.GetServicesWithLabels(context.Background(), namespace, groupParam)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(HttpError("could not get service information: %v\n", err))
		return
	}

	response, err := json.Marshal(&HttpResponse{
		Data:    services,
		Success: true,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(HttpError("could not marshal response data: %v\n", err))
		return
	}

	w.Write(response)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("healthy"))
}
