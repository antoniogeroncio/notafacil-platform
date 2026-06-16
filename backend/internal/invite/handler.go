package invite

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/notafacil/platform/backend/internal/user"
	"github.com/notafacil/platform/backend/pkg/httpx"
)

// Handler is the HTTP entry point for team invitations (Controller layer).
type Handler struct {
	svc *Service
}

// NewHandler builds an invite Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type createRequest struct {
	Email string    `json:"email"`
	Role  user.Role `json:"role"`
}

// Create handles POST /api/v1/invites (Admin only). It maps domain errors to
// the contract's HTTP statuses.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var body createRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.Error(w, http.StatusUnprocessableEntity, "validation_error", "Corpo da requisição inválido.")
		return
	}

	inv, err := h.svc.CreateInvite(r.Context(), CreateInput{Email: body.Email, Role: body.Role})
	switch {
	case errors.Is(err, ErrValidation):
		httpx.Error(w, http.StatusUnprocessableEntity, "validation_error", "E-mail ou papel inválidos.")
		return
	case errors.Is(err, ErrEmailConflict):
		httpx.Error(w, http.StatusConflict, "email_conflict", "Este e-mail já está cadastrado na plataforma.")
		return
	case err != nil:
		httpx.Error(w, http.StatusInternalServerError, "internal_error", "Não foi possível processar o convite.")
		return
	}

	httpx.JSON(w, http.StatusCreated, map[string]any{"invite": inv.ToView()})
}
