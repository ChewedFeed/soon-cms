package retro

import (
	"encoding/json"
	"github.com/bugfixes/go-bugfixes/logs"
	"io/ioutil"
	"net/http"
)

func jsonError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	if err := json.NewEncoder(w).Encode(struct {
		Error string `json:"error"`
	}{
		Error: err.Error(),
	}); err != nil {
		logs.Debugf("jsonError: %v", err)
	}
}

func (c CMS) ServicesHandler(w http.ResponseWriter, r *http.Request) {
	services, err := c.getServices()
	if err != nil {
		logs.Infof("ServicesHandler: %v", err)
		jsonError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(services); err != nil {
		logs.Debugf("ServicesHandler: %v", err)
	}
}

func (c CMS) ServiceHandler(w http.ResponseWriter, r *http.Request) {
	service, err := c.getService(r.PathValue("service"))
	if err != nil {
		logs.Infof("ServiceHandler: %v", err)
		jsonError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(service); err != nil {
		logs.Debugf("ServiceHandler: %v", err)
	}
}

func (c CMS) ScriptHandler(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadFile("script.js")
	if err != nil {
		logs.Infof("ScriptHandler: %v", err)
		jsonError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/javascript")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(b)
	if err != nil {
		logs.Infof("ScriptHandler write: %v", err)
		jsonError(w, err)
		return
	}
}
