package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lordbaldwin1/chirpy/internal/auth"
	"github.com/lordbaldwin1/chirpy/internal/database"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	log.Printf("bearerToken: %s: \n", bearerToken)
	userId, err := auth.ValidateJWT(bearerToken, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var params parameters
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode request body JSON", err)
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long", fmt.Errorf("error: chirp must be less than 140 characters"))
		return
	}
	params.Body = removeProfanity(params.Body)

	chirp, err := cfg.queries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   params.Body,
		UserID: userId,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create chirp in database", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}

// would be better using a map!
func removeProfanity(text string) string {
	profanities := [3]string{"kerfuffle", "sharbert", "fornax"}

	textArr := strings.Split(text, " ")

	for i, word := range textArr {
		for _, profanity := range profanities {
			if strings.ToLower(word) == profanity {
				textArr[i] = "****"
			}
		}
	}

	return strings.Join(textArr, " ")
}
