package api

import (
	"log/slog"
	"net/http"
	application "proofofconcept/internal/application/smoke"
	domainsmoke "proofofconcept/internal/domain/smoke"
)

type SmokeAPI struct {
	smokeApp *application.SmokeApp
	logger   *slog.Logger
}

func NewSmokeAPI(smokeApp *application.SmokeApp) *SmokeAPI {
	return &SmokeAPI{
		smokeApp: smokeApp,
	}
}

func (apu *SmokeAPI) Hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello There"))
}

func (api *SmokeAPI) NewSmoke(w http.ResponseWriter, r *http.Request) {
	// input will receive data from the Request body
	var input struct {
		Description string `json:"description"`
	}

	// read the json body into input
	err := readJSON(w, r, &input)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// create a domain type
	//--------------------------

	smoke, err := domainsmoke.NewSmoke(input.Description)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// save to the database
	// --------------------

	smoke, err = api.smokeApp.Create(r.Context(), smoke)
	if err != nil {
		api.logger.Error("create error", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// reply to client
	err = writeJSON(w, http.StatusCreated, map[string]any{"smoke": smoke}, nil)
	if err != nil {
		api.logger.Error("writing to client", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
