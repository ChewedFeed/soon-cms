package retro

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"github.com/bugfixes/go-bugfixes/logs"
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

func jsonResponse(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		logs.Debugf("jsonResponse: %v", err)
	}
}

func (c CMS) CreateServiceHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, err)
		return
	}

	svc, err := c.createService(req)
	if err != nil {
		jsonError(w, err)
		return
	}
	jsonResponse(w, http.StatusCreated, svc)
}

func (c CMS) UpdateServiceHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, err)
		return
	}

	svc, err := c.updateService(r.PathValue("service"), req)
	if err != nil {
		jsonError(w, err)
		return
	}
	jsonResponse(w, http.StatusOK, svc)
}

func (c CMS) DeleteServiceHandler(w http.ResponseWriter, r *http.Request) {
	if err := c.deleteService(r.PathValue("service")); err != nil {
		jsonError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c CMS) CreateLinkHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, err)
		return
	}

	link, err := c.createLink(r.PathValue("service"), req)
	if err != nil {
		jsonError(w, err)
		return
	}
	jsonResponse(w, http.StatusCreated, link)
}

func (c CMS) DeleteLinkHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		jsonError(w, err)
		return
	}

	if err := c.deleteLink(id); err != nil {
		jsonError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c CMS) CreateRoadmapHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateRoadmapRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, err)
		return
	}

	item, err := c.createRoadmapItem(r.PathValue("service"), req)
	if err != nil {
		jsonError(w, err)
		return
	}
	jsonResponse(w, http.StatusCreated, item)
}

func (c CMS) UpdateRoadmapHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		jsonError(w, err)
		return
	}

	var req UpdateRoadmapRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, err)
		return
	}

	item, err := c.updateRoadmapItem(id, req)
	if err != nil {
		jsonError(w, err)
		return
	}
	jsonResponse(w, http.StatusOK, item)
}

func (c CMS) DeleteRoadmapHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		jsonError(w, err)
		return
	}

	if err := c.deleteRoadmapItem(id); err != nil {
		jsonError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c CMS) ListTasksHandler(w http.ResponseWriter, r *http.Request) {
	tasks, err := c.getLaunchTasks(r.PathValue("service"))
	if err != nil {
		jsonError(w, err)
		return
	}
	jsonResponse(w, http.StatusOK, tasks)
}

func (c CMS) CreateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Completed bool `json:"completed"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, err)
		return
	}

	task, err := c.createLaunchTask(r.PathValue("service"), req.Completed)
	if err != nil {
		jsonError(w, err)
		return
	}
	jsonResponse(w, http.StatusCreated, task)
}

func (c CMS) UpdateTaskHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		jsonError(w, err)
		return
	}

	var req struct {
		Completed bool `json:"completed"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, err)
		return
	}

	task, err := c.updateLaunchTask(id, req.Completed)
	if err != nil {
		jsonError(w, err)
		return
	}
	jsonResponse(w, http.StatusOK, task)
}

func (c CMS) DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		jsonError(w, err)
		return
	}

	if err := c.deleteLaunchTask(id); err != nil {
		jsonError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c CMS) ScriptHandler(w http.ResponseWriter, r *http.Request) {
	b, err := os.ReadFile("script.js")
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
