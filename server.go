package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"pulseservice/config"
	"pulseservice/ingestionhandlers"
	"pulseservice/requesthandlers"
	"syscall"
	"time"

	"sync"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

//"github.com/gorilla/handlers"
type myServer struct {
	http.Server
	shutdownReq chan bool
	reqCount    uint32
	WG          sync.WaitGroup
}

// NewServer - this is the init function for the server process
func NewServer(port string) *myServer {

	//create server - this version creates a server that listens on any address
	s := &myServer{
		Server: http.Server{
			Addr:         "127.0.0.1:" + port,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
		shutdownReq: make(chan bool),
	}

	router := mux.NewRouter()

	//register handlers
	router.HandleFunc("/pulse", s.RootHandler)

	// Swagger
	sh := http.StripPrefix("/pulse/V01/swaggerui/", http.FileServer(http.Dir("./swaggerui/")))
	router.PathPrefix("/pulse/V01/swaggerui/").Handler(sh)

	// Main endpoints
	router.HandleFunc("/pulse/V01/hello", ingestionhandlers.HelloHandler).Methods("POST")
	router.HandleFunc("/pulse/V01/servicestatus", ingestionhandlers.ServiceStatusHandler).Methods("POST")
	router.HandleFunc("/pulse/V01/servicealert", ingestionhandlers.ServiceAlertHandler).Methods("POST")

	router.HandleFunc("/pulse/V01/latest", requesthandlers.ServiceGetLatestHandler).Methods("GET")
	router.HandleFunc("/pulse/V01/latestforapp", requesthandlers.ServiceGetLatestMessageForAppHandler).Methods("GET")
	router.HandleFunc("/pulse/V01/lateststatus", requesthandlers.GetLatestServiceStatusMessagesHandler).Methods("GET")
	router.HandleFunc("/pulse/V01/latesthello", requesthandlers.GetLatestHelloMessagesHandler).Methods("GET")
	router.HandleFunc("/pulse/V01/messages", requesthandlers.ServiceGetMessageForAppBetweenTimesHandler).Methods("GET")
	//router.HandleFunc("/pulse/V01/alert", requesthandlers.ServiceGetAlertHandler).Methods("GET")
	//router.HandleFunc("/pulse/V01/tokenlogin", requesthandlers.TokenLogin).Methods("GET")

	// CORS stuff
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "X-API-KEY", "X-Request-Token", "Content-Type"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"})
	s.Handler = handlers.CORS(headersOk, originsOk, methodsOk)(router)

	return s
}

func (s *myServer) WaitShutdown() {
	irqSig := make(chan os.Signal, 1)
	signal.Notify(irqSig, syscall.SIGINT, syscall.SIGTERM)

	//Wait interrupt or shutdown request through /shutdown
	select {
	case sig := <-irqSig:
		log.Printf("Shutdown request (signal: %v)", sig)
	case sig := <-s.shutdownReq:
		log.Printf("Shutdown request (/shutdown %v)", sig)
	}
	log.Printf("Stopping API server ...")
	close(config.Done)
	//Create shutdown context with 10 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//shutdown the server
	err := s.Shutdown(ctx)
	if err != nil {
		log.Printf("Shutdown request error: %v", err)
	}
	fmt.Println("Waiting for waitgroup to clear")
	log.Println("Waiting for waitgroup to clear")

	s.WG.Wait()

}

func (s *myServer) RootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Pulse - see /pulse/V01/swaggerui/ for documentation\n"))
}
