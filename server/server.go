package server

import (
	"context"
	"crypto/tls"
	"errors"
	"log"
	"net"
	"net/http"
	"time"
)

// Server holds two servers: HTTP and HTTPS
type Server struct {
	httpAddr    string
	httpsAddr   string
	handler     http.Handler
	httpServer  *http.Server
	httpsServer *http.Server
	certMap     map[string]tls.Certificate
	certs       []tls.Certificate
}

// New returns a new server
func New(handler http.Handler, httpAddr, httpsAddr string) (*Server, error) {
	srv := &Server{
		httpAddr:  httpAddr,
		httpsAddr: httpsAddr,
		handler:   handler,
		certMap:   make(map[string]tls.Certificate),
	}

	return srv, nil
}

// AddCertificate adds a certificate
func (srv *Server) AddCertificate(id, cert, key string) error {
	crt, err := tls.X509KeyPair([]byte(cert), []byte(key))
	if err != nil {
		return err
	}
	srv.certMap[id] = crt
	srv.certMapToSlice()
	return nil
}

// RemoveCertificate removes a certificate
func (srv *Server) RemoveCertificate(id string) error {
	if _, ok := srv.certMap[id]; !ok {
		return errors.New("certificate doesn't exist")
	}
	delete(srv.certMap, id)
	srv.certMapToSlice()
	return nil
}

func (srv *Server) certMapToSlice() {
	slice := make([]tls.Certificate, len(srv.certMap))
	i := 0
	for _, crt := range srv.certMap {
		slice[i] = crt
		i++
	}
	srv.certs = slice
}

// ListenAndServeHTTP starts the HTTP server
func (srv *Server) ListenAndServeHTTP() error {
	if srv.httpServer == nil {
		srv.httpServer = &http.Server{
			Addr:    srv.httpAddr,
			Handler: srv.handler,
		}
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		srv.httpServer.Shutdown(ctx)
		cancel()
	}
	ln, err := net.Listen("tcp", srv.httpAddr)
	if err != nil {
		return err
	}
	go srv.httpServer.Serve(ln)
	return nil
}

// ListenAndServeHTTPS (re)starts the HTTPS server
func (srv *Server) ListenAndServeHTTPS() error {
	if len(srv.certs) == 0 {
		log.Print("no certificate available")
		return nil
	}
	config := &tls.Config{}
	config.Certificates = srv.certs
	config.BuildNameToCertificate()
	if srv.httpsServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		srv.httpsServer.Shutdown(ctx)
		cancel()
		srv.httpsServer = nil
	}
	ln, err := net.Listen("tcp", srv.httpsAddr)
	if err != nil {
		return err
	}
	tlsListener := tls.NewListener(ln, config)
	srv.httpsServer = &http.Server{
		Addr:    srv.httpAddr,
		Handler: srv.handler,
	}
	go srv.httpsServer.Serve(tlsListener)
	return nil
}
