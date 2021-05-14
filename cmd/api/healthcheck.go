package main

import (
	"fmt"
	"net/http"
)

func (app *application) healthcheck(w http.ResponseWriter, r *http.Request) {
	app.logger.LogInfo(fmt.Sprintf("%s - %s: %s", r.RemoteAddr, r.Method, r.URL.String()), nil)
	d := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     version,
		},
	}

	err := app.writeJSON(w, http.StatusOK, d, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
