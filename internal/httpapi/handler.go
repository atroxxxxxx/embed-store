package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	database "github.com/atroxxxxxx/embed-store/internal/db"
	"go.uber.org/zap"
)

type Repo interface {
	// InsertChunk inserts chunk to db
	InsertChunk(ctx context.Context, chunk *database.Chunk) (int64, error)
	// ChunkByID searches chunk by ID in db
	ChunkByID(ctx context.Context, id int64) (*database.Chunk, error)
}

type Handler struct {
	db     Repo
	logger *zap.Logger
}

var (
	ErrNullArgs = errors.New("null constructor arguments")
)

func New(db Repo, logger *zap.Logger) (*Handler, error) {
	if db == nil || logger == nil {
		return nil, ErrNullArgs
	}

	return &Handler{
		db:     db,
		logger: logger,
	}, nil
}

func (obj *Handler) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/chunks", obj.post)
	mux.HandleFunc("/chunks/", obj.get)
	return mux
}

func (obj *Handler) sendErrResponse(writer http.ResponseWriter, message string, errCode int, err error) {
	if err != nil {
		obj.logger.Error(message, zap.Int("code", errCode), zap.Error(err))
	} else {
		obj.logger.Error(message, zap.Int("code", errCode))
	}

	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(errCode)
	_, _ = writer.Write([]byte(message))
}

func (obj *Handler) post(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		obj.sendErrResponse(writer, "method not allowed", http.StatusMethodNotAllowed, nil)
		return
	}
	var req Request

	dec := json.NewDecoder(request.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(&req)
	if err != nil {
		obj.sendErrResponse(writer, "bad request", http.StatusBadRequest, err)
		return
	}
	chunk, err := Map(&req)
	if err != nil {
		obj.sendErrResponse(writer, "bad request", http.StatusBadRequest, err)
		return
	}
	id, err := obj.db.InsertChunk(request.Context(), chunk)
	if err != nil {
		obj.sendErrResponse(writer, "internal server error", http.StatusInternalServerError, err)
		return
	}
	obj.logger.Debug("successfully inserted", zap.Int64("id", id))
	resp, err := Unmap(chunk, false)
	if err != nil {
		obj.sendErrResponse(writer, "internal server error", http.StatusInternalServerError, err)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(writer).Encode(resp)
	if err != nil {
		obj.logger.Warn("encode internal server error",
			zap.Int("code", http.StatusInternalServerError),
			zap.Error(err))
	}
	obj.logger.Info("data successfully added", zap.Int64("id", id))
}

func (obj *Handler) get(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		obj.sendErrResponse(writer, "method not allowed", http.StatusMethodNotAllowed, nil)
		return
	}
	rawID, found := strings.CutPrefix(request.URL.Path, "/chunks/")
	if !found || rawID == "" {
		obj.sendErrResponse(writer, "not found", http.StatusNotFound, nil)
		return
	}
	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			obj.sendErrResponse(writer, "not found", http.StatusNotFound, err)
		} else {
			obj.sendErrResponse(writer, "bad request", http.StatusBadRequest, err)
		}
		return
	}
	chunk, err := obj.db.ChunkByID(request.Context(), id)
	if err != nil {
		obj.sendErrResponse(writer, "not found", http.StatusNotFound, err)
		return
	}
	query := request.URL.Query()
	embed := query.Get("embed") == "1"
	resp, err := Unmap(chunk, embed)
	if err != nil {
		obj.sendErrResponse(writer, "internal server error", http.StatusInternalServerError, err)
		return
	}

	writer.Header().Set("Location", fmt.Sprintf("/chunks/%d", id))
	writer.WriteHeader(http.StatusOK)
	err = json.NewEncoder(writer).Encode(resp)
	if err != nil {
		obj.logger.Warn("encode internal server error",
			zap.Int("code", http.StatusInternalServerError),
			zap.Error(err))
	}
	obj.logger.Info("chunk found", zap.Int64("id", resp.ID))
}
