package handler

import "net/http"

func Heathz(writer http.ResponseWriter, rawRequest *http.Request) {

	writer.Header().Set("Content-Type", "text/plain")
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write([]byte("ok"))
}

type ReadinessCheck func(rawRequest *http.Request) error

func Readyz(checks ...ReadinessCheck) http.HandlerFunc {

	return func(writer http.ResponseWriter, rawRequest *http.Request) {

		for _, check := range checks {

			if err := check(rawRequest); err != nil {

				http.Error(writer, "not ready:"+err.Error(), http.StatusServiceUnavailable)

				return
			}
		}

		writer.Header().Set("Content-Type", "text/plain")
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte("ready"))
	}
}
