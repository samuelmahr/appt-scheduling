package controllers

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func respondError(ctx context.Context, w http.ResponseWriter, status int, message string, causer error) {
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]interface{}{
		"error": message,
	}

	if status >= 400 {
		log.WithFields(log.Fields{
			"message": message,
			"causer":  causer,
		},
		).Error("oops")
	}

	if typer, ok := causer.(errorTyper); ok {
		resp["type"] = typer.ErrorType()
	}

	if errors.Cause(causer) == sql.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.WriteHeader(status)
	}

	bytes, _ := json.Marshal(resp)
	_, _ = w.Write(bytes)
}

func respondModel(ctx context.Context, w http.ResponseWriter, status int, model interface{}) {
	b, err := json.Marshal(model)
	if err != nil {
		respondError(ctx, w, 500, "error generating response", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(b)
}
