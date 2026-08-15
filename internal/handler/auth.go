package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/SunMaungOo/task-runner-auth/internal/service"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	AccessToken string `json:"accessToken"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (handler *AuthHandler) Register(writer http.ResponseWriter, rawRequest *http.Request) {

	var request registerRequest

	if err := json.NewDecoder(rawRequest.Body).Decode(&request); err != nil {

		writeJSON(writer, http.StatusBadRequest, errorResponse{Error: "invalid request body"})

		return
	}

	user, err := handler.svc.Register(rawRequest.Context(), request.Email, request.Password)

	if err != nil {

		switch {
		case errors.Is(err, service.ErrorInvalidEmail), errors.Is(err, service.ErrorWeakPassword):
			writeJSON(writer, http.StatusBadRequest, errorResponse{Error: err.Error()})

		case errors.Is(err, service.ErrorEmailTaken):
			writeJSON(writer, http.StatusConflict, errorResponse{Error: err.Error()})

		default:
			writeJSON(writer, http.StatusInternalServerError, errorResponse{Error: "internal error"})

		}

		return

	}

	writeJSON(writer, http.StatusCreated, registerResponse{
		ID:    user.ID,
		Email: user.Email})

}

func (handler *AuthHandler) Login(writer http.ResponseWriter, rawRequest *http.Request) {

	var request loginRequest

	if err := json.NewDecoder(rawRequest.Body).Decode(&request); err != nil {

		writeJSON(writer, http.StatusBadRequest, errorResponse{Error: "invalid request body"})

		return
	}

	token, err := handler.svc.Login(rawRequest.Context(), request.Email, request.Password)

	if err != nil {

		if errors.Is(err, service.ErrorInvalidCredential) {

			writeJSON(writer, http.StatusUnauthorized, errorResponse{Error: err.Error()})

			return
		}

		writeJSON(writer, http.StatusInternalServerError, errorResponse{Error: "internal error"})

		return
	}

	writeJSON(writer, http.StatusOK, loginResponse{
		AccessToken: token,
	})
}

func writeJSON(writer http.ResponseWriter, status int, value any) {

	writer.Header().Set("Content-Type", "application/json")

	writer.WriteHeader(status)

	_ = json.NewEncoder(writer).Encode(value)
}
