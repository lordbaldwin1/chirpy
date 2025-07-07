package main

import (
	"database/sql"
	"net/http"
	"sort"

	"github.com/google/uuid"
	"github.com/lordbaldwin1/chirpy/internal/database"
)

func (cfg *apiConfig) handlerChirpsGet(w http.ResponseWriter, r *http.Request) {
	authorID := r.URL.Query().Get("author_id")
	sortOrder := r.URL.Query().Get("sort")
	var dbChirps []database.Chirp
	var err error

	if authorID == "" {
		dbChirps, err = cfg.queries.GetChirps(r.Context())
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't get chirps from DB", err)
			return
		}
	} else {
		authorUUID, err := uuid.Parse(authorID)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Couldn't parse authorID", err)
		}
		dbChirps, err = cfg.queries.GetChirpsByAuthorID(r.Context(), authorUUID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't get chirps from DB", err)
			return
		}
	}

	var chirps []Chirp
	for _, dbChirp := range dbChirps {
		chirps = append(chirps, Chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			Body:      dbChirp.Body,
			UserID:    dbChirp.UserID,
		})
	}

	if sortOrder == "asc" {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].CreatedAt.Before(chirps[j].CreatedAt)
		})
	}
	if sortOrder == "desc" {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
		})
	}

	respondWithJSON(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) handlerChirpsGetByID(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("chirpID")

	dbChirp, err := cfg.queries.GetChirpsByID(r.Context(), uuid.MustParse(chirpID))
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Chirp not found", nil)
			return
		}

		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirp", err)
		return
	}
	respondWithJSON(w, http.StatusOK, Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	})
}
