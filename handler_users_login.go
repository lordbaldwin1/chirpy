package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/lordbaldwin1/chirpy/internal/auth"
	"github.com/lordbaldwin1/chirpy/internal/database"
)

func (cfg *apiConfig) handlerUsersLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type response struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	decoder := json.NewDecoder(r.Body)
	var params parameters
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode request body into JSON", err)
		return
	}

	user, err := cfg.queries.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't find user", err)
		return
	}

	err = auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect password", err)
		return
	}

	accessToken, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to make JWT token", err)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't make refresh token", err)
		return
	}

	err = cfg.queries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().AddDate(0, 0, 60),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't store refresh token in db", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		},
		Token:        accessToken,
		RefreshToken: refreshToken,
	})
}
