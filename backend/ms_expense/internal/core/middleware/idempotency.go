package middleware

import (
	"bytes"
	"context"
	"net/http"
	"shared/apierror"
	"shared/auth/contexts"
	"shared/cache"
	"time"
	"uuid"
)

type IdempotencyStore interface {
	Find(ctx context.Context, key uuid.UUID) (*StoredResponse, error)
	Save(ctx context.Context, key, userID uuid.UUID, endpoint string, status int, body []byte) error
}

type StoredResponse struct {
	Status int
	Body   []byte
}

type errorHandler interface {
	HandlerError(w http.ResponseWriter, r *http.Request, err error)
}

type Idempotency struct {
	store      IdempotencyStore
	cache      cache.Cache
	errHandler errorHandler
	lockTTL    time.Duration
}

func NewIdempotency(
	store IdempotencyStore,
	c cache.Cache,
	errHandler errorHandler,
) *Idempotency {
	return &Idempotency{
		store:      store,
		cache:      c,
		errHandler: errHandler,
		lockTTL:    10 * time.Second,
	}
}

func (m *Idempotency) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := r.Header.Get("Idempotency-Key")
		if raw == "" {
			next.ServeHTTP(w, r)
			return
		}

		key, err := uuid.Parse(raw)
		if err != nil {
			m.errHandler.HandlerError(w, r, apierror.NewBadRequestError(err))
			return
		}

		user := contexts.GetUser(r.Context())
		if user.IsAnonymous() {
			m.errHandler.HandlerError(w, r, apierror.ErrAuthenticationRequired)
			return
		}
		userID := user.GetID()

		stored, err := m.store.Find(r.Context(), key)
		if err != nil {
			m.errHandler.HandlerError(w, r, err)
			return
		}
		if stored != nil {
			writeStored(w, stored)
			return
		}

		lockKey := "idempotency:lock:" + key.String()
		acquired, err := m.cache.SetNX(r.Context(), lockKey, "1", &m.lockTTL)
		if err != nil {
			m.errHandler.HandlerError(w, r, err)
			return
		}

		if !acquired {
			w.Header().Set("Retry-After", "1")
			m.errHandler.HandlerError(w, r, apierror.NewHTTPError(
				"concurrent request with same Idempotency-Key",
				http.StatusConflict,
				nil,
			))
			return
		}

		defer func() {
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_ = m.cache.Delete(cleanupCtx, lockKey)
		}()

		rec := newResponseRecorder(w)
		next.ServeHTTP(rec, r)

		if rec.status >= 200 && rec.status < 300 {
			endpoint := r.Method + " " + r.URL.Path
			_ = m.store.Save(r.Context(), key, userID, endpoint, rec.status, rec.body.Bytes())
		}
	})
}

func writeStored(w http.ResponseWriter, s *StoredResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Idempotent-Replay", "true")
	w.WriteHeader(s.Status)
	_, _ = w.Write(s.Body)
}

type responseRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
	body        bytes.Buffer
}

func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{ResponseWriter: w, status: http.StatusOK}
}

func (r *responseRecorder) WriteHeader(status int) {
	if r.wroteHeader {
		return
	}
	r.status = status
	r.wroteHeader = true
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func (r *responseRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}
