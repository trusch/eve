package handler

import (
	"log"
	"net/http"

	loadbalancer "github.com/trusch/yap/loadbalancer/manager"
	middleware "github.com/trusch/yap/middleware/manager"
)

// Handler is the global http request handler
type Handler struct {
	LBManager *loadbalancer.Manager
	MWManager *middleware.Manager
}

func (handler *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	chain, err := handler.MWManager.BuildChain(req, handler.LBManager)
	if err != nil {
		log.Printf("Error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	chain.ServeHTTP(w, req)
}

// New returns a new handler
func New() *Handler {
	lbManager := loadbalancer.New()
	mwManager := middleware.New()
	return &Handler{lbManager, mwManager}
}
