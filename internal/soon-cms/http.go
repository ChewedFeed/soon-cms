package retro

import (
	"encoding/json"
	"net/http"

	bugLog "github.com/bugfixes/go-bugfixes/logs"
	"github.com/go-chi/chi/v5"
)

func jsonError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	if err := json.NewEncoder(w).Encode(struct {
		Error string `json:"error"`
	}{
		Error: err.Error(),
	}); err != nil {
		bugLog.Local().Debugf("jsonError: %v", err)
	}
}

func (c CMS) ServicesHandler(w http.ResponseWriter, r *http.Request) {
	services, err := c.getServices()
	if err != nil {
		jsonError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(services); err != nil {
		bugLog.Local().Debugf("ServicesHandler: %v", err)
	}
}

func (c CMS) ServiceHandler(w http.ResponseWriter, r *http.Request) {
	service, err := c.getService(chi.URLParam(r, "service"))
	if err != nil {
		jsonError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(service); err != nil {
		bugLog.Local().Debugf("ServiceHandler: %v", err)
	}
}
