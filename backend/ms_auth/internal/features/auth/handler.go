package auth

import (
	"context"
	"net/http"
	"uuid"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"shared/apierror"
	"shared/auth/contexts"
	"shared/auth/cookies"
	"shared/httpx/httpjson"
	"shared/httpx/httputils"
)

type authHandler struct {
	service    service
	errHandler errorHandler
}

type errorHandler interface {
	HandlerError(w http.ResponseWriter, r *http.Request, err error)
}

type service interface {
	Authenticate(ctx context.Context, email, password string) (*TokenResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	ValidateAccessToken(token string) bool
}

func newHandler(
	service service,
	errHandler errorHandler,
) *authHandler {
	return &authHandler{
		service:    service,
		errHandler: errHandler,
	}
}

func (h *authHandler) Authenticate(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("ms_auth/internal/features/auth")
	ctx, span := tracer.Start(r.Context(), "authHandler.Authenticate")
	defer span.End()

	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := httpjson.ReadJSON(w, r, &input); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to read JSON")
		h.errHandler.HandlerError(w, r, apierror.NewHTTPError(
			err.Error(),
			http.StatusBadRequest,
			err,
		))
		return
	}

	span.SetAttributes(attribute.String("user.email", input.Email))

	token, err := h.service.Authenticate(ctx, input.Email, input.Password)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Authentication failed")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	cookies.SetAccessTokenCookie(w, token.AccessToken, token.ExpiresIn)
	cookies.SetRefreshTokenCookie(w, token.RefreshToken, token.RefreshExpiresIn)

	span.SetStatus(codes.Ok, "Authentication successful")
	httputils.Respond(w, r, http.StatusOK, map[string]bool{"ok": true}, nil, h.errHandler)
}

func (h *authHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("ms_auth/internal/features/auth")
	ctx, span := tracer.Start(r.Context(), "authHandler.RefreshToken")
	defer span.End()

	refreshToken := cookies.RefreshTokenFromRequest(r)

	if refreshToken == "" {
		h.errHandler.HandlerError(w, r, apierror.NewHTTPError(
			"missing refresh token",
			http.StatusUnauthorized,
			nil,
		))
		return
	}

	token, err := h.service.RefreshToken(ctx, refreshToken)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Refresh token failed")
		h.errHandler.HandlerError(w, r, err)
		return
	}

	cookies.SetAccessTokenCookie(w, token.AccessToken, token.ExpiresIn)
	cookies.SetRefreshTokenCookie(w, token.RefreshToken, token.RefreshExpiresIn)

	span.SetStatus(codes.Ok, "Token refreshed")
	httputils.Respond(w, r, http.StatusOK, map[string]bool{"ok": true}, nil, h.errHandler)
}

func (h *authHandler) Logout(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("ms_auth/internal/features/auth")
	ctx, span := tracer.Start(r.Context(), "authHandler.Logout")
	defer span.End()

	refreshToken := cookies.RefreshTokenFromRequest(r)
	if refreshToken != "" {
		_ = h.service.Logout(ctx, refreshToken)
	}

	cookies.ClearAuthCookies(w)

	span.SetStatus(codes.Ok, "Logout successful")
	httputils.Respond(w, r, http.StatusNoContent, nil, nil, h.errHandler)
}

func (h *authHandler) Session(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("ms_auth/internal/features/auth")
	_, span := tracer.Start(r.Context(), "authHandler.Session")
	defer span.End()

	user := contexts.GetUser(r.Context())

	if user == nil || user.GetID() == uuid.Nil() {
		httputils.Respond(w, r, http.StatusOK,
			SessionResponse{Authenticated: false}, nil, h.errHandler)
		return
	}

	token := cookies.TokenFromRequest(r)
	authenticated := token != "" && h.service.ValidateAccessToken(token)
	if !authenticated {
		httputils.Respond(w, r, http.StatusOK,
			SessionResponse{Authenticated: authenticated}, nil, h.errHandler)
		return
	}

	id := user.GetID()
	email := user.GetUsername()

	span.SetStatus(codes.Ok, "Session checked")
	httputils.Respond(w, r, http.StatusOK, SessionResponse{
		Authenticated: authenticated,
		UserID:        id,
		Email:         &email,
	}, nil, h.errHandler)
}
