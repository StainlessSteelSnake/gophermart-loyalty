package handlers

import (
	"log"
	"net/http"

	"github.com/StainlessSteelSnake/gophermart-loyalty/internal/auth"
	"github.com/StainlessSteelSnake/gophermart-loyalty/internal/orders"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	*chi.Mux
	authenticator    auth.Authenticator
	orders           orders.OrderAdderGetter
	currentUserLogin string
	baseURL          string
}

func NewHandler(baseURL string, a auth.Authenticator, o orders.OrderAdderGetter) *Handler {
	log.Println("Base URL:", baseURL)

	handler := &Handler{
		Mux:           chi.NewMux(),
		authenticator: a,
		orders:        o,
		baseURL:       baseURL,
	}

	handler.Route("/", func(r chi.Router) {
		handler.Use(handler.authenticate)
		handler.Use(gzipHandler)

		r.Post("/api/user/register", handler.registerUser)
		r.Post("/api/user/login", handler.loginUser)
		r.Post("/api/user/orders", handler.addOrder)
		r.Get("/api/user/orders", handler.getOrders)
		r.Get("/api/user/balance", handler.getBalance)
		r.Post("/api/user/balance/withdraw", handler.withdrawPoints)
		r.Get("/api/user/withdrawals", handler.getWithdrawals)
		r.MethodNotAllowed(handler.badRequest)
	})

	return handler
}

func (h *Handler) badRequest(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "неподдерживаемый запрос: '"+r.RequestURI+"'", http.StatusBadRequest)
}
