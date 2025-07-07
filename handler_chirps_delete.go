package main

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/lordbaldwin1/chirpy/internal/auth"
	"github.com/lordbaldwin1/chirpy/internal/database"
)

func (cfg *apiConfig) handlerChirpsDelete(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("chirpID")
	chirpUUID, err := uuid.Parse(chirpID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID", err)
		return
	}

	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "No bearer token in auth header", err)
		return
	}

	userID, err := auth.ValidateJWT(accessToken, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid JWT token", err)
		return
	}

	dbChirp, err := cfg.queries.GetChirpsByID(r.Context(), chirpUUID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp not found", err)
		return
	}
	if dbChirp.UserID != userID {
		respondWithError(w, http.StatusForbidden, "User doesn't have access to this chirp", err)
		return
	}

	err = cfg.queries.DeleteChirp(r.Context(), database.DeleteChirpParams{
		UserID: userID,
		ID:     chirpUUID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete chirp", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
