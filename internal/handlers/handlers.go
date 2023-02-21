package handlers

import (
	"github.com/StainlessSteelSnake/gophermart-loyalty/internal/auth"
	"log"
	"net/http"

	"github.com/StainlessSteelSnake/gophermart-loyalty/internal/database"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	*chi.Mux
	storage          database.Storager
	auth             auth.Authenticator
	currentUserLogin string
	baseURL          string
}

func NewHandler(s database.Storager, baseURL string, a auth.Authenticator) *Handler {
	log.Println("Base URL:", baseURL)

	handler := &Handler{
		Mux:     chi.NewMux(),
		storage: s,
		auth:    a,
		baseURL: baseURL,
	}

	handler.Route("/", func(r chi.Router) {
		handler.Use(handler.authenticate)
		//handler.Use(gzipHandler)

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

func (h *Handler) addOrder(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) getOrders(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) getBalance(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) withdrawPoints(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) getWithdrawals(w http.ResponseWriter, r *http.Request) {

}
