package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"cosmasgithinji.net/simplesnippetbox/pkg/models"
	"github.com/justinas/nosurf"
)

// write an error message and stacktrace to the errroLog
// send a generic 500 Internal Server Error to the user
func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())

	app.errorLog.Output(2, trace)

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// send a specific status code and corresponding description to user
// e.g 400 Bad Request for BadRequest
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// wrapper around clientError. Send 404 Not Found to user
func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

func (app *application) addDefaultData(td *templateData, r *http.Request) *templateData {
	if td == nil {
		td = &templateData{}
	}
	td.CSRFToken = nosurf.Token(r)
	td.AuthenticatedUser = app.authenticatedUser(r)
	td.CurrentYear = time.Now().Year()
	td.Flash = app.session.PopString(r, "flash") // add flash message if one exists
	return td
}

func (app *application) render(w http.ResponseWriter, r *http.Request, name string, td *templateData) {
	ts, ok := app.templateCache[name] //Retrieve template from cache
	if !ok {
		app.serverError(w, fmt.Errorf("the template %s does not exist", name))
		return
	}

	buf := new(bytes.Buffer) // init a new buffer

	err := ts.Execute(buf, app.addDefaultData(td, r)) // write to buffer instead of http.ResponseWriter
	if err != nil {
		app.serverError(w, err)
	}

	buf.WriteTo(w) //Write buffer contents to http.ResponseWriter
}

// Retriev user details from context
func (app *application) authenticatedUser(r *http.Request) *models.User {
	user, ok := r.Context().Value(contextKeyUser).(*models.User)
	if !ok {
		return nil
	}
	return user
}

// func getEnv(key string) (string, error) {
// 	value := os.Getenv(key)
// 	if value == "" {
// 		return "", fmt.Errorf("environment variable %s is not set", key)
// 	}
// 	return value, nil
// }

func getFlagOrEnv(flagValue *string, envVar string) (string, error) {
	if *flagValue != "" { // If the flag is provided, use it
		return *flagValue, nil
	}
	// Otherwise, fall back to the env variable
	envValue := os.Getenv(envVar)
	if envValue == "" {
		return "", fmt.Errorf("neither flag nor environment variable %s is set", envVar)
	}
	return envValue, nil
}
