package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type returnVal struct {
		CleanedBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	chrp := parameters{}
	err := decoder.Decode(&chrp)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode body: ", err)
		return
	}

	if len(chrp.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long", fmt.Errorf("error: chirp must be less than 140 characters"))
		return
	}

	// now clean the input of profanity
	chrp.Body = removeProfanity(chrp.Body)
	respondWithJSON(w, 200, returnVal{
		CleanedBody: chrp.Body,
	})
}

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
