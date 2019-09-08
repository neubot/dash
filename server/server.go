// Package server implements the DASH server
package server

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/neubot/dash/common"
	"github.com/neubot/dash/internal"
	"github.com/neubot/dash/mockable"
)

type sessionInfo struct {
	iteration    int64
	serverSchema common.ServerSchema
	stamp        time.Time
}

// Handler is the DASH handler
type Handler struct {
	// Datadir is the directory where to save measurements
	Datadir string

	// Dependencies contains mockable dependencies. Usually you don't
	// want to touch them unless you're into unit testing.
	Dependencies mockable.Dependencies

	// Logger is the logger to use. This field is initialized by the
	// NewHandler constructor to a do-nothing logger.
	Logger common.Logger

	// MaxIterations is the maximum number of iterations that this
	// server is allowing a clients to perform.
	MaxIterations int64

	mtx      sync.Mutex
	sessions map[string]*sessionInfo
	stop     chan interface{}
}

// NewHandler creates a new handler instance
func NewHandler(datadir string) *Handler {
	return &Handler{
		Datadir:       datadir,
		Dependencies:  mockable.NewDependencies(),
		Logger:        internal.NoLogger{},
		MaxIterations: 17,
		sessions:      make(map[string]*sessionInfo),
		stop:          make(chan interface{}),
	}
}

func (h *Handler) createSession(UUID string) {
	now := time.Now()
	session := &sessionInfo{
		stamp: now,
		serverSchema: common.ServerSchema{
			ServerSchemaVersion: common.CurrentServerSchemaVersion,
			ServerTimestamp:     now.Unix(),
		},
	}
	h.mtx.Lock()
	defer h.mtx.Unlock()
	h.sessions[UUID] = session
}

type sessionState int

const (
	sessionMissing = sessionState(iota)
	sessionActive
	sessionExpired
)

func (h *Handler) getSessionState(UUID string) sessionState {
	h.mtx.Lock()
	defer h.mtx.Unlock()
	session, ok := h.sessions[UUID]
	if !ok {
		return sessionMissing
	}
	if session.iteration >= h.MaxIterations {
		return sessionExpired
	}
	return sessionActive
}

func (h *Handler) updateSession(UUID string, count int) {
	now := time.Now()
	h.mtx.Lock()
	defer h.mtx.Unlock()
	session, ok := h.sessions[UUID]
	if ok {
		session.serverSchema.Server = append(
			session.serverSchema.Server, common.ServerResults{
				Iteration: session.iteration,
				Ticks:     now.Sub(session.stamp).Seconds(),
				Timestamp: now.Unix(),
			},
		)
		session.iteration++
	}
}

func (h *Handler) popSession(UUID string) *sessionInfo {
	h.mtx.Lock()
	defer h.mtx.Unlock()
	session, ok := h.sessions[UUID]
	if ok == false {
		return nil
	}
	delete(h.sessions, UUID)
	return session
}

// CountSessions counts the number of open sessions.
func (h *Handler) CountSessions() (count int) {
	h.mtx.Lock()
	defer h.mtx.Unlock()
	count = len(h.sessions)
	return
}

func (h *Handler) reapStaleSessions() {
	h.mtx.Lock()
	defer h.mtx.Unlock()
	h.Logger.Debugf("reaper: inspecting %d sessions", len(h.sessions))
	now := time.Now()
	var stale []string
	for UUID, session := range h.sessions {
		const toomuch = 60 * time.Second
		if now.Sub(session.stamp) > toomuch {
			stale = append(stale, UUID)
		}
	}
	h.Logger.Debugf("reaper: reaping %d stale sessions", len(stale))
	for _, UUID := range stale {
		delete(h.sessions, UUID)
	}
}

func (h *Handler) negotiate(w http.ResponseWriter, r *http.Request) {
	address, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	UUID, err := h.Dependencies.UUIDNewRandom()
	if err != nil {
		w.WriteHeader(500)
		return
	}
	// Implementation note: we do not include any vector of speeds
	// in the response, meaning that the client should use its predefined
	// vector of speeds rather than using ours. This vector of speeds
	// thing is bad anyway, because clients may not upgrade. To escape
	// from this limitation, we use a different strategy in this code
	// where we pick any client chosen value within a specific range.
	//
	// A side effect of this implementation choice is that we are now
	// tolerating incoming requests that do not contain any body.
	data, err := h.Dependencies.JSONMarshal(common.NegotiateResponse{
		Authorization: UUID.String(),
		QueuePos:      0,
		RealAddress:   address,
		Unchoked:      1,
	})
	if err != nil {
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(data); err == nil {
		h.createSession(UUID.String())
	}
}

const (
	// MinSize is the minimum segment size that this server can return.
	//
	// The client requests two second chunks. The minimum emulated streaming
	// speed is the minimum streaming speed (in kbit/s) multiplied by 1000
	// to obtain bit/s, divided by 8 to obtain bytes/s and multiplied by the
	// two seconds to obtain the minimum segment size.
	MinSize = 100 * 1000 / 8 * 2

	// MaxSize is the maximum segment size that this server can return. See
	// the docs of MinSize for more information on how it is computed.
	MaxSize = 20000 * 1000 / 8 * 2

	authorization = "Authorization"
)

var (
	once          sync.Once
	minSizeString = fmt.Sprintf("%d", MinSize)
)

func (h *Handler) genbody(count *int) (data []byte, err error) {
	// Implementation note: because one may be lax during refactoring
	// and may end up using count rather than len(data) and because
	// count may be way bigger than the real data length, I've changed
	// this function to _also_ update count to the real value.
	once.Do(func() {
		rand.Seed(time.Now().UTC().UnixNano())
	})
	if *count < MinSize {
		*count = MinSize
	}
	if *count > MaxSize {
		*count = MaxSize
	}
	data = make([]byte, *count)
	_, err = h.Dependencies.RandRead(data)
	return
}

func (h *Handler) download(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Header.Get(authorization)
	state := h.getSessionState(sessionID)
	if state == sessionMissing {
		w.WriteHeader(400)
		return
	}
	// The Neubot implementation used to raise runtime error in this case
	// leading to 500 being returned to the client. Here we deviate from
	// the original implementation returning a value that seems to be much
	// more useful and actionable to the client.
	if state == sessionExpired {
		w.WriteHeader(429)
		return
	}
	siz := strings.Replace(r.URL.Path, "/dash/download", "", -1)
	if strings.HasPrefix(siz, "/") {
		siz = siz[1:]
	}
	if siz == "" {
		siz = minSizeString
	}
	count, err := strconv.Atoi(siz)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	data, err := h.genbody(&count)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	h.updateSession(sessionID, len(data))
	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.Write(data)
}

type resultsFile struct {
	writer *gzip.Writer
	fp     *os.File
}

func (h *Handler) savedata(session *sessionInfo) error {
	name := path.Join(h.Datadir, session.stamp.Format("2006/01/02"))
	err := h.Dependencies.OSMkdirAll(name, 0755)
	if err != nil {
		return err
	}
	name += "/neubot-dash-" + session.stamp.Format("20060102T150405.000000000Z") + ".json.gz"
	// My assumption here is that we have nanosecond precision and hence it's
	// unlikely to have conflicts. If I'm wrong, O_EXCL will let us know.
	filep, err := h.Dependencies.OSOpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}
	defer filep.Close()
	zipper, err := h.Dependencies.GzipNewWriterLevel(filep, gzip.BestSpeed)
	if err != nil {
		return err
	}
	defer zipper.Close()
	data, err := h.Dependencies.JSONMarshal(session.serverSchema)
	if err != nil {
		return err
	}
	_, err = zipper.Write(data)
	return err
}

func (h *Handler) collect(w http.ResponseWriter, r *http.Request) {
	session := h.popSession(r.Header.Get(authorization))
	if session == nil {
		w.WriteHeader(400)
		return
	}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	err = json.Unmarshal(data, &session.serverSchema.Client)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	data, err = h.Dependencies.JSONMarshal(session.serverSchema.Server)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	err = h.savedata(session)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.Write([]byte(data))
}

// RegisterHandlers registers handlers for the URLs used by the DASH
// experiment. The following prefixes are registered:
//
// - /negotiate/dash
// - /dash/download
// - /collect/dash
//
// The /negotiate/dash prefix is used to create a measurement
// context for a dash client. The /download/dash prefix is
// used by clients to request data segments. The /collect/dash
// prefix is used to submit client measurements.
func (h *Handler) RegisterHandlers(mux *http.ServeMux) {
	mux.HandleFunc(common.NegotiatePath, h.negotiate)
	mux.HandleFunc(common.DownloadPath, h.download)
	mux.HandleFunc(common.DownloadPathNoTrailingSlash, h.download)
	mux.HandleFunc(common.CollectPath, h.collect)
}

func (h *Handler) reaperLoop(ctx context.Context) {
	defer close(h.stop)
	for ctx.Err() == nil {
		const reapInterval = 14 * time.Second
		time.Sleep(reapInterval)
		h.reapStaleSessions()
	}
}

// JoinReaper blocks until the reaper has terminated
func (h *Handler) JoinReaper() {
	<-h.stop
}

// StartReaper starts the reaper goroutine that makes sure that
// we write back results of incomplete measurements. This goroutine
// will terminate when the |ctx| context becomes expired.
func (h *Handler) StartReaper(ctx context.Context) {
	go h.reaperLoop(ctx)
}
