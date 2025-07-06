package main

import (
	"net/http"
	"time"

	"github.com/lordbaldwin1/chirpy/internal/auth"
	"github.com/lordbaldwin1/chirpy/internal/database"
)

func (cfg *apiConfig) handleRefreshToken(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}

	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't get bearer token to refresh", err)
		return
	}

	user, err := cfg.queries.GetUserFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find user", err)
		return
	}

	accessToken, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't make new refresh token", err)
		return
	}

	err = cfg.queries.UpdateRefreshToken(r.Context(), database.UpdateRefreshTokenParams{
		Token:  accessToken,
		UserID: user.ID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update refresh token in db", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		Token: accessToken,
	})
}
