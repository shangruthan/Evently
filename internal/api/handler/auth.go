package handler

import (
	"encoding/json"
	"errors"
	"evently/internal/service"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgconn"
)

type AuthHandler struct {
	authService *service.AuthService
	log         *slog.Logger
}

func NewAuthHandler(authService *service.AuthService, log *slog.Logger) *AuthHandler {
	return &AuthHandler{authService: authService, log: log}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload")
		return
	}

	user, err := h.authService.Register(r.Context(), input.Name, input.Email, input.Password)
	if err != nil {
		h.log.Error("Failure during user registration", "error", err) // THIS IS THE NEW DEBUG LOG

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			RespondWithError(w, http.StatusConflict, "email_exists", "A user with this email already exists")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "server_error", "Could not create user")
		return
	}

	RespondWithJSON(w, http.StatusCreated, user)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload")
		return
	}

	token, err := h.authService.Login(r.Context(), input.Email, input.Password)
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, "unauthorized", err.Error())
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]string{"token": token})
}
