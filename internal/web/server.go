package web

import (
	"net/http"
	"time"

	"github.com/didip/tollbooth/v6"
	"github.com/didip/tollbooth/v6/limiter"
	"github.com/gorilla/mux"
)

type ListenParams struct {
	// Address is server listen address (socket)
	Address string

	// ReadTimeout is request read timeout
	ReadTimeout time.Duration

	// WriteTimeout is response write timeout
	WriteTimeout time.Duration

	// LimitExpirationTTL is token bucket ttl
	LimitExpirationTTL time.Duration

	// ClientRPSQuota is request per second quota for each client (by IP)
	ClientRPSQuota float64
}

func (p ListenParams) limiterOpts() *limiter.ExpirableOptions {
	if p.LimitExpirationTTL > 0 {
		return &limiter.ExpirableOptions{DefaultExpirationTTL: p.LimitExpirationTTL}
	}

	return nil
}

type Server struct {
	*http.Server
	Router *mux.Router
}

// NewServer constructs new HTTP server with specified params
func NewServer(p ListenParams) *Server {
	router := mux.NewRouter()

	router.Path("aa").Subrouter().Use()
	httpSrv := &http.Server{
		Handler:      router,
		Addr:         p.Address,
		WriteTimeout: p.WriteTimeout,
		ReadTimeout:  p.ReadTimeout,
	}

	// Add rate-limiter middleware only if client request quota is defined.
	if p.ClientRPSQuota > 0 {
		rlimit := tollbooth.NewLimiter(p.ClientRPSQuota, p.limiterOpts())
		httpSrv.Handler = tollbooth.LimitHandler(rlimit, router)
	}

	return &Server{
		Server: httpSrv,
		Router: router,
	}
}
