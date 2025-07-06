package main

import (
	"net/http"

	"github.com/lordbaldwin1/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't get bearer token while revoking", err)
		return
	}

	err = cfg.queries.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't revoke refresh token from db", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
