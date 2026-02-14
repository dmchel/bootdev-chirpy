package validation

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/dmchel/bootdev-chirpy/utils"
)

func ValidateChirpHandler(w http.ResponseWriter, r *http.Request) {
	type chirp struct {
		Body string `json:"body"`
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	type result struct {
		CleanedBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	ch := chirp{}

	err := decoder.Decode(&ch)
	if err != nil {
		log.Println("Failed to decode chirp:", err)
		utils.InternalServerErrorHandler(w, r)
		return
	}

	if len(ch.Body) <= 140 {
		var response result
		response.CleanedBody = cleanChirp(ch.Body)
		utils.Respond(w, 200, response)
	} else {
		var response errorResponse
		response.Error = "Chirp is too long"
		utils.Respond(w, 400, response)
	}
}

func cleanChirp(body string) string {
	if len(body) == 0 {
		return ""
	}

	nastyWords := [3]string{"kerfuffle", "sharbert", "fornax"}
	words := strings.Split(body, " ")
	for i, word := range words {
		lw := strings.ToLower(word)
		for _, nword := range nastyWords {
			if lw == nword {
				words[i] = "****"
				break
			}
		}
	}

	return strings.Join(words, " ")
}
