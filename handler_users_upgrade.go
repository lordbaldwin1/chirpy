package main

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/lordbaldwin1/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerUsersUpgrade(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}

	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Failed to get API key", err)
		return
	}

	if apiKey != cfg.polkaAPIKey {
		respondWithError(w, http.StatusUnauthorized, "Incorrect Polka API key", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var params parameters
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to decode request body into JSON", err)
		return
	}

	if params.Event != "user.upgraded" {
		respondWithError(w, http.StatusNoContent, "Event not user.upgraded", err)
		return
	}
	userUUID, err := uuid.Parse(params.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to parse userID", err)
		return
	}

	_, err = cfg.queries.UpgradeUserToChirpyRed(r.Context(), userUUID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't find user to upgrade", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
