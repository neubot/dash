package server

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
)

// ListenAndServeTLS is like http.ListenAndServeTLS except that it
// wraps *tls.Conn so that we can access the underlying net.Conn
func ListenAndServeTLS(addr, certFile, keyFile string, handler http.Handler) error {
	// Implementation note: this code cannot call the internal function
	// to setup serving HTTP/2 over TLS, hence we're more limited that
	// what we can actually do inside the standard library. Yet, for DASH
	// this is fine and it would theoretically also be fine for the NDT
	// server because there we use WebSocket.
	server := newserver(addr, handler)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer listener.Close()
	config := new(tls.Config)
	config.Certificates = make([]tls.Certificate, 1)
	config.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}
	return server.Serve(newdashlistener(listener, config))
}

// ListenAndServe is like http.ListenAndServe except that it
// allows us to access the underlying conn.
func ListenAndServe(addr string, handler http.Handler) error {
	server := newserver(addr, handler)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return server.Serve(newdashlistener(listener, nil))
}

func newserver(addr string, handler http.Handler) (server *http.Server) {
	server = &http.Server{
		Addr:    addr,
		Handler: handler,
		ConnContext: func(ctx context.Context, conn net.Conn) context.Context {
			return withConn(ctx, conn)
		},
	}
	return
}

type dashlistener struct {
	net.Listener
	config *tls.Config
}

func newdashlistener(inner net.Listener, config *tls.Config) *dashlistener {
	return &dashlistener{Listener: inner, config: config}
}

func (dl *dashlistener) Accept() (net.Conn, error) {
	underlying, err := dl.Listener.Accept()
	if err != nil {
		return nil, err
	}
	conn := underlying
	if dl.config != nil {
		conn = tls.Server(underlying, dl.config)
	}
	return &dashconn{
		Conn:       conn,
		underlying: underlying,
	}, nil
}

type dashconn struct {
	net.Conn
	underlying net.Conn
}

func (c *dashconn) Underlying() net.Conn {
	return c.underlying
}

type contextkey struct{}

func withConn(ctx context.Context, conn net.Conn) context.Context {
	return context.WithValue(ctx, contextkey{}, conn)
}

func contextConn(ctx context.Context) (conn net.Conn) {
	conn, _ = ctx.Value(contextkey{}).(net.Conn)
	return
}
