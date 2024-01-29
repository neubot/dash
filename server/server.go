// Package server implements the DASH server.
package server

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/neubot/dash/internal"
	"github.com/neubot/dash/model"
	"github.com/neubot/dash/spec"
)

// sessionInfo contains information about an active session.
type sessionInfo struct {
	// iteration is the number of iterations done by the active session.
	iteration int64

	// serverSchema contains the server schema for the given session.
	serverSchema model.ServerSchema

	// stamp is when we created this struct.
	stamp time.Time
}

// dependencies abstracts the dependencies used by [*Handler].
type dependencies struct {
	GzipNewWriterLevel func(w io.Writer, level int) (*gzip.Writer, error)
	IOReadAll          func(r io.Reader) ([]byte, error)
	JSONMarshal        func(v interface{}) ([]byte, error)
	OSMkdirAll         func(path string, perm os.FileMode) error
	OSOpenFile         func(name string, flag int, perm os.FileMode) (*os.File, error)
	RandRead           func(p []byte) (n int, err error)
	Savedata           func(session *sessionInfo) error
	UUIDNewRandom      func() (uuid.UUID, error)
}

// Handler is the DASH handler. Please use NewHandler to construct
// a valid instance of this type (the zero value is invalid).
//
// You need to call the RegisterHandlers method to register the proper
// DASH handlers. You also need to call StartReaper to periodically
// get rid of sessions that have been running for too much. If you don't
// call StartReaper, you will eventually run out of RAM.
type Handler struct {
	// Datadir is the directory where to save measurements.
	Datadir string

	// Logger is the logger to use. This field is initialized by the
	// NewHandler constructor to a do-nothing logger.
	Logger model.Logger

	// deps contains the [*Handler] dependencies.
	deps dependencies

	// maxIterations is the maximum allowed number of iterations.
	maxIterations int64

	// mtx protects the sessions map.
	mtx sync.Mutex

	// sessions maps a session UUID to session info.
	sessions map[string]*sessionInfo

	// stop is closed when the reaper goroutine is stopped.
	stop chan any
}

// NewHandler creates a new [*Handler] instance.
func NewHandler(datadir string) *Handler {
	handler := &Handler{
		Datadir:       datadir,
		Logger:        internal.NoLogger{},
		deps:          dependencies{}, // initialized later
		maxIterations: 17,
		mtx:           sync.Mutex{},
		sessions:      make(map[string]*sessionInfo),
		stop:          make(chan interface{}),
	}
	handler.deps = dependencies{
		GzipNewWriterLevel: gzip.NewWriterLevel,
		IOReadAll:          io.ReadAll,
		JSONMarshal:        json.Marshal,
		OSMkdirAll:         os.MkdirAll,
		OSOpenFile:         os.OpenFile,
		RandRead:           rand.Read, // math/rand is okay to use here
		Savedata:           handler.savedata,
		UUIDNewRandom:      uuid.NewRandom,
	}
	return handler
}

// createSession creates a session using the given UUID.
//
// This method LOCKS and MUTATES the .sessions field.
func (h *Handler) createSession(UUID string) {
	now := time.Now()
	session := &sessionInfo{
		stamp: now,
		serverSchema: model.ServerSchema{
			ServerSchemaVersion: spec.CurrentServerSchemaVersion,
			ServerTimestamp:     now.Unix(),
		},
	}
	h.mtx.Lock()
	defer h.mtx.Unlock()
	h.sessions[UUID] = session
}

// sessionState is the state of a measurement session.
type sessionState int

const (
	// sessionMissing indicates that a session with the given UUID does not exist.
	sessionMissing = sessionState(iota)

	// sessionActive indicates that a session exists and it has not performed
	// the maximum number of allowed iterations yet.
	sessionActive

	// sessionExpired is a session that performed all the possible iterations.
	sessionExpired
)

// getSessionState returns the state of the session with the given UUID.
func (h *Handler) getSessionState(UUID string) sessionState {
	h.mtx.Lock()
	defer h.mtx.Unlock()
	session, ok := h.sessions[UUID]
	if !ok {
		return sessionMissing
	}
	if session.iteration >= h.maxIterations {
		return sessionExpired
	}
	return sessionActive
}

// updateSession updates the state of the session with the given UUID after
// we successfully performed a new iteration.
//
// When the UUID maps to an existing session, this method SAFELY MUTATES the
// session's serverSchema by adding a new measurement result and by
// incrementing the number of iterations.
//
// The integer argument, currently ignored, contains the number of bytes
// that were sent as part of the current DASH iteration.
func (h *Handler) updateSession(UUID string, _ int) {
	now := time.Now()
	h.mtx.Lock()
	defer h.mtx.Unlock()
	session, ok := h.sessions[UUID]
	if ok {
		session.serverSchema.Server = append(
			session.serverSchema.Server, model.ServerResults{
				Iteration: session.iteration,
				Ticks:     now.Sub(session.stamp).Seconds(),
				Timestamp: now.Unix(),
			},
		)
		session.iteration++
	}
}

// popSession returns nil if a session with the given UUID does not exist, otherwise
// is SAFELY REMOVES and returns the corresponding [*sessionInfo].
func (h *Handler) popSession(UUID string) *sessionInfo {
	h.mtx.Lock()
	defer h.mtx.Unlock()
	session, ok := h.sessions[UUID]
	if !ok {
		return nil
	}
	delete(h.sessions, UUID)
	return session
}

// CountSessions SAFELY COUNTS and returns the number of active sessions.
func (h *Handler) CountSessions() (count int) {
	h.mtx.Lock()
	defer h.mtx.Unlock()
	count = len(h.sessions)
	return
}

// reapStaleSessions SAFELY REMOVES all the sessions created more than 60 seconds ago.
func (h *Handler) reapStaleSessions() {
	h.mtx.Lock()
	defer h.mtx.Unlock()
	h.Logger.Debugf("reapStaleSessions: inspecting %d sessions", len(h.sessions))
	now := time.Now()
	var stale []string
	for UUID, session := range h.sessions {
		const toomuch = 60 * time.Second
		if now.Sub(session.stamp) > toomuch {
			stale = append(stale, UUID)
		}
	}
	h.Logger.Debugf("reapStaleSessions: reaping %d stale sessions", len(stale))
	for _, UUID := range stale {
		delete(h.sessions, UUID)
	}
}

// negotiate implements the /negotiate/dash handler.
//
// Neubot originally implemented access control and parameters negotiation in
// this preliminary measurement stage. This implementation relies on m-lab's locate
// service to implement access control so we only negotiate the parameters. We
// assume that m-lab's incoming request interceptor will take care of the authorization
// token passed as part of the request URL.
//
// This method SAFELY MUTATES the sessions map by creating a new session UUID. If
// clients do not call this method first, measurements will fail for lack of a valid
// session UUID.
func (h *Handler) negotiate(w http.ResponseWriter, r *http.Request) {
	// Obtain the client's remote address.
	address, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		h.Logger.Warnf("negotiate: net.SplitHostPort: %s", err.Error())
		w.WriteHeader(500)
		return
	}

	// Create a new random UUID for the session.
	//
	// We assume we're not going to have UUID conflicts.
	UUID, err := h.deps.UUIDNewRandom()
	if err != nil {
		h.Logger.Warnf("negotiate: uuid.NewRandom: %s", err.Error())
		w.WriteHeader(500)
		return
	}

	// Prepare the response.
	//
	// Implementation note: we do not include any vector of speeds
	// in the response, meaning that the client should use its predefined
	// vector of speeds rather than using ours. This vector of speeds
	// thing is bad anyway, because clients may not upgrade. To escape
	// from this limitation, we use a different strategy in this code
	// where we pick any client chosen value within a specific range.
	//
	// A side effect of this implementation choice is that we are now
	// tolerating incoming requests that do not contain any body.
	data, err := h.deps.JSONMarshal(model.NegotiateResponse{
		Authorization: UUID.String(),
		QueuePos:      0,
		RealAddress:   address,
		Unchoked:      1,
	})

	// Make sure we can properly marshal the response.
	if err != nil {
		h.Logger.Warnf("negotiate: json.Marshal: %s", err.Error())
		w.WriteHeader(500)
		return
	}

	// Send the response.
	w.Header().Set("Content-Type", "application/json")
	h.createSession(UUID.String())
	_, _ = w.Write(data)
}

const (
	// minSize is the minimum segment size that this server can return.
	//
	// The client requests two second chunks. The minimum emulated streaming
	// speed is the minimum streaming speed (in kbit/s) multiplied by 1000
	// to obtain bit/s, divided by 8 to obtain bytes/s and multiplied by the
	// two seconds to obtain the minimum segment size.
	minSize = 100 * 1000 / 8 * 2

	// maxSize is the maximum segment size that this server can return. See
	// the docs of MinSize for more information on how it is computed.
	maxSize = 30000 * 1000 / 8 * 2

	// authorization is the key for the Authorization header.
	authorization = "Authorization"
)

// minSize string is the string representation of the minSize constant.
var minSizeString = fmt.Sprintf("%d", minSize)

// genbody generates the body and updates the count argument to
// be within the acceptable bounds allowed by the protocol.
//
// Implementation note: because one may be lax during refactoring
// and may end up using count rather than len(data) and because
// count may be way bigger than the real data length, I've changed
// this function to _also_ update count to the real value.
func (h *Handler) genbody(count *int) (data []byte, err error) {
	if *count < minSize {
		*count = minSize
	}
	if *count > maxSize {
		*count = maxSize
	}
	data = make([]byte, *count)
	_, err = h.deps.RandRead(data)
	return
}

// download implements the /dash/download handler.
func (h *Handler) download(w http.ResponseWriter, r *http.Request) {
	// make sure we have a valid session
	sessionID := r.Header.Get(authorization)
	state := h.getSessionState(sessionID)
	if state == sessionMissing {
		h.Logger.Warn("download: session missing")
		w.WriteHeader(400)
		return
	}

	// Make sure the session did not expire (i.e., that it did not
	// send too many requests as part of the same session).
	//
	// The Neubot implementation used to raise runtime error in this case
	// leading to 500 being returned to the client. Here we deviate from
	// the original implementation returning a value that seems to be much
	// more useful and actionable to the client.
	if state == sessionExpired {
		h.Logger.Warn("download: session expired")
		w.WriteHeader(429)
		return
	}

	// obtain the number of bytes we should send to the client according
	// to what the client would like to receive.
	siz := strings.Replace(r.URL.Path, "/dash/download", "", -1)
	siz = strings.TrimPrefix(siz, "/")
	if siz == "" {
		siz = minSizeString
	}
	count, err := strconv.Atoi(siz)
	if err != nil {
		h.Logger.Warnf("download: strconv.Atoi: %s", err.Error())
		w.WriteHeader(400)
		return
	}

	// generate body possibly adjusting the count if it falls out of
	// the acceptable bounds for the response size.
	data, err := h.genbody(&count)
	if err != nil {
		h.Logger.Warnf("download: genbody: %s", err.Error())
		w.WriteHeader(500)
		return
	}

	// Register that the session has done an iteration.
	h.updateSession(sessionID, len(data))

	// Send the response.
	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	_, _ = w.Write(data)
}

// savedata is an utility function saving information about this session.
func (h *Handler) savedata(session *sessionInfo) error {
	// obtain the directory path where to write
	name := path.Join(h.Datadir, "dash", session.stamp.Format("2006/01/02"))

	// make sure we have the correct directory hierarchy
	err := h.deps.OSMkdirAll(name, 0755)
	if err != nil {
		h.Logger.Warnf("savedata: os.MkdirAll: %s", err.Error())
		return err
	}

	// append the file name to the path
	//
	// TODO(bassosimone): this code does not work as intended on Windows
	name += "/neubot-dash-" + session.stamp.Format("20060102T150405.000000000Z") + ".json.gz"

	// open the results file
	//
	// My assumption here is that we have nanosecond precision and hence it's
	// unlikely to have conflicts. If I'm wrong, O_EXCL will let us know.
	filep, err := h.deps.OSOpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		h.Logger.Warnf("savedata: os.OpenFile: %s", err.Error())
		return err
	}
	defer filep.Close()

	// wrap the output file with a gzipper
	zipper, err := h.deps.GzipNewWriterLevel(filep, gzip.BestSpeed)
	if err != nil {
		h.Logger.Warnf("savedata: gzip.NewWriterLevel: %s", err.Error())
		return err
	}
	defer zipper.Close()

	// marshal the measurement to JSON
	data, err := h.deps.JSONMarshal(session.serverSchema)
	if err != nil {
		h.Logger.Warnf("savedata: json.Marshal: %s", err.Error())
		return err
	}

	// write compressed data into the file
	_, err = zipper.Write(data)
	return err
}

// collect implements the /collect/dash handler.
func (h *Handler) collect(w http.ResponseWriter, r *http.Request) {
	// make sure we have a session
	session := h.popSession(r.Header.Get(authorization))
	if session == nil {
		h.Logger.Warn("collect: session missing")
		w.WriteHeader(400)
		return
	}

	// read the incoming measurements collected by the client
	data, err := h.deps.IOReadAll(r.Body)
	if err != nil {
		h.Logger.Warnf("collect: ioutil.ReadAll: %s", err.Error())
		w.WriteHeader(400)
		return
	}

	// unmarshal client data from JSON into the server data structure
	err = json.Unmarshal(data, &session.serverSchema.Client)
	if err != nil {
		h.Logger.Warnf("collect: json.Unmarshal: %s", err.Error())
		w.WriteHeader(400)
		return
	}

	// serialize all
	data, err = h.deps.JSONMarshal(session.serverSchema.Server)
	if err != nil {
		h.Logger.Warnf("collect: json.Marshal: %s", err.Error())
		w.WriteHeader(500)
		return
	}

	// save on disk
	err = h.deps.Savedata(session)
	if err != nil {
		// Error already printed by h.savedata()
		w.WriteHeader(500)
		return
	}

	// tell the client we're all good
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	_, _ = w.Write([]byte(data))
}

// RegisterHandlers registers handlers for the URLs used by the DASH
// experiment. The following prefixes are registered:
//
// - /negotiate/dash
// - /dash/download/{size}
// - /collect/dash
//
// The /negotiate/dash prefix is used to create a measurement
// context for a dash client. The /download/dash prefix is
// used by clients to request data segments. The /collect/dash
// prefix is used to submit client measurements.
//
// For historical reasons /dash/download is an alias for
// using the /dash/download/ prefix.
func (h *Handler) RegisterHandlers(mux *http.ServeMux) {
	mux.HandleFunc(spec.NegotiatePath, h.negotiate)
	mux.HandleFunc(spec.DownloadPath, h.download)
	mux.HandleFunc(spec.DownloadPathNoTrailingSlash, h.download)
	mux.HandleFunc(spec.CollectPath, h.collect)
}

// reaperLoop is the goroutine that periodically reaps expired sessions.
func (h *Handler) reaperLoop(ctx context.Context) {
	h.Logger.Debug("reaperLoop: start")
	defer h.Logger.Debug("reaperLoop: done")
	defer close(h.stop)
	for ctx.Err() == nil {
		const reapInterval = 14 * time.Second
		time.Sleep(reapInterval)
		h.reapStaleSessions()
	}
}

// StartReaper starts the reaper goroutine that makes sure that
// we write back results of incomplete measurements. This goroutine
// will terminate when the |ctx| context becomes expired.
func (h *Handler) StartReaper(ctx context.Context) {
	go h.reaperLoop(ctx)
}

// JoinReaper blocks until the reaper has terminated
func (h *Handler) JoinReaper() {
	<-h.stop
}
