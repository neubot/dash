// Package server implements the DASH server
package server

import (
	"compress/gzip"
	"context"
	"encoding/json"
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

	"github.com/google/uuid"
	"github.com/neubot/dash/common"
)

type resultsRecord struct {
	iteration    int64
	serverSchema common.ServerSchema
	stamp        time.Time
}

// Handler is the DASH handler
type Handler struct {
	// Datadir is the directory where to save measurements
	Datadir string

	mtx     sync.Mutex
	records map[string]*resultsRecord
}

// NewHandler creates a new handler instance
func NewHandler(datadir string) *Handler {
	return &Handler{
		Datadir: datadir,
		records: make(map[string]*resultsRecord),
	}
}

func (h *Handler) createRecord(UUID string) {
	now := time.Now()
	record := &resultsRecord{
		stamp: now,
		serverSchema: common.ServerSchema{
			ServerSchemaVersion: common.CurrentServerSchemaVersion,
			ServerTimestamp:     now.Unix(),
		},
	}
	h.mtx.Lock()
	defer h.mtx.Unlock()
	h.records[UUID] = record
}

func (h *Handler) updateRecord(UUID string, count int) (ok bool) {
	now := time.Now()
	h.mtx.Lock()
	defer h.mtx.Unlock()
	record, ok := h.records[UUID]
	if ok {
		record.serverSchema.Server = append(
			record.serverSchema.Server, common.ServerResults{
				Iteration: record.iteration,
				Ticks:     now.Sub(record.stamp).Seconds(),
				Timestamp: now.Unix(),
			},
		)
		record.iteration++
	}
	return
}

func (h *Handler) popRecord(UUID string) *resultsRecord {
	h.mtx.Lock()
	defer h.mtx.Unlock()
	record, ok := h.records[UUID]
	if ok == false {
		return nil
	}
	delete(h.records, UUID)
	return record
}

func (h *Handler) reapStaleRecords() {
	h.mtx.Lock()
	defer h.mtx.Unlock()
	now := time.Now()
	var stale []string
	for UUID, record := range h.records {
		const toomuch = 60 * time.Second
		if now.Sub(record.stamp) > toomuch {
			stale = append(stale, UUID)
		}
	}
	for _, UUID := range stale {
		delete(h.records, UUID)
	}
}

func (h *Handler) negotiate(w http.ResponseWriter, r *http.Request) {
	address, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	UUID, err := uuid.NewRandom()
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
	data, err := json.Marshal(common.NegotiateResponse{
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
		h.createRecord(UUID.String())
	}
}

const authorization = "Authorization"

var once sync.Once

func (h *Handler) download(w http.ResponseWriter, r *http.Request) {
	// The client requests two second chunks. The minimum emulated streaming
	// speed is 100 kbit/s. The maximum is 20,000 kbit/s. Here we use as
	// minimum and maximum the conversion of such values in bytes. Everything
	// that is outside this range is coerced in this range.
	const (
		minSize       = 100 * 1000 / 8 * 2
		minSizeString = string(minSize)
		maxSize       = 20000 * 1000 / 8 * 2
	)
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
	if count < minSize {
		count = minSize
	}
	if count > maxSize {
		count = maxSize
	}
	if h.updateRecord(r.Header.Get(authorization), count) == false {
		w.WriteHeader(400)
		return
	}
	data := make([]byte, count)
	once.Do(func() {
		rand.Seed(time.Now().UTC().UnixNano())
	})
	_, err = rand.Read(data)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.Write(data)
}

type resultsFile struct {
	writer *gzip.Writer
	fp     *os.File
}

func (h *Handler) savedata(record *resultsRecord) error {
	name := path.Join(h.Datadir, record.stamp.Format("2006/01/02"))
	err := os.MkdirAll(name, 0755)
	if err != nil {
		return err
	}
	name += "/neubot-dash-" + record.stamp.Format("20060102T150405.000000000Z") + ".json.gz"
	// My assumption here is that we have nanosecond precision and hence it's
	// unlikely to have conflicts. If I'm wrong, O_EXCL will let us know.
	filep, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}
	defer filep.Close()
	zipper, err := gzip.NewWriterLevel(filep, gzip.BestSpeed)
	if err != nil {
		return err
	}
	defer zipper.Close()
	data, err := json.Marshal(record.serverSchema)
	if err != nil {
		return err
	}
	_, err = zipper.Write(data)
	return err
}

func (h *Handler) collect(w http.ResponseWriter, r *http.Request) {
	record := h.popRecord(r.Header.Get(authorization))
	if record == nil {
		w.WriteHeader(400)
		return
	}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	err = json.Unmarshal(data, &record.serverSchema.Client)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	data, err = json.Marshal(record.serverSchema.Server)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	err = h.savedata(record)
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
	mux.HandleFunc(common.CollectPath, h.collect)
}

func (h *Handler) reaperLoop(ctx context.Context) {
	for ctx.Err() == nil {
		const reapInterval = 14 * time.Second
		time.Sleep(reapInterval)
		h.reapStaleRecords()
	}
}

// StartReaper starts the reaper goroutine that makes sure that
// we write back results of incomplete measurements. This goroutine
// will terminate when the |ctx| context becomes expired.
func (h *Handler) StartReaper(ctx context.Context) {
	go h.reaperLoop(ctx)
}
