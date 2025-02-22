package main

import "net/http"


func (app *application) healthCheckHandler(w http.ResponseWriter , r *http.Request){
	app.jsonResponse(w,http.StatusOK,"it's working")
}