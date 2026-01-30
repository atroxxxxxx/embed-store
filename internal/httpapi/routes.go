package httpapi

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	database "github.com/atroxxxxxx/embed-store/internal/db"
	"github.com/pgvector/pgvector-go"
	"go.uber.org/zap"
)

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
		if errors.Is(err, database.ErrDuplicateKey) {
			obj.sendErrResponse(writer, "conflict", http.StatusPaymentRequired, err)
		} else {
			obj.sendErrResponse(writer, "internal server error", http.StatusInternalServerError, err)
		}
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

func (obj *Handler) search(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		obj.sendErrResponse(writer, "method not allowed", http.StatusMethodNotAllowed, ErrInvalidType)
		return
	}

	var req SearchRequest
	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		obj.sendErrResponse(writer, "bad request", http.StatusBadRequest, err)
		return
	}
	if len(req.Embedding) != database.VectorSize {
		obj.sendErrResponse(writer,
			"bad request: invalid embedding length", http.StatusBadRequest, ErrInvalidEmbeddingLen,
		)
		return
	}
	if req.Limit <= 0 {
		req.Limit = 4
	} else if req.Limit > 100 {
		req.Limit = 100
	}

	vec := pgvector.NewVector(req.Embedding)
	chunks, err := obj.db.Search(request.Context(), &vec, req.Limit)
	if err != nil {
		obj.sendErrResponse(writer, "internal server error", http.StatusInternalServerError, err)
		return
	}

	responses := make([]*Response, 0, len(chunks))
	for _, chunk := range chunks {
		resp, err := Unmap(chunk, req.IncludeEmbedding)
		if err != nil {
			obj.sendErrResponse(writer, "internal server error", http.StatusInternalServerError, err)
		}
		responses = append(responses, &resp)
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(writer).Encode(responses); err != nil {
		obj.logger.Warn("encode response failed", zap.Error(err))
	}
}
