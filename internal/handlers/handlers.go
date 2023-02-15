package handlers

import (
	"log"
	"net/http"

	"github.com/StainlessSteelSnake/gophermart-loyalty/internal/database"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	*chi.Mux
	storage database.Storager
	baseURL string
}

func NewHandler(s database.Storager, baseURL string) *Handler {
	log.Println("Base URL:", baseURL)

	handler := &Handler{
		chi.NewMux(),
		s,
		baseURL,
	}

	handler.Route("/", func(r chi.Router) {
		//handler.Use(handler.auth.Authenticate)
		//handler.Use(gzipHandler)

		r.Post("/api/user/register", handler.registerUser)
		r.Post("/api/user/login", handler.authenticateUser)
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

func (h *Handler) registerUser(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) authenticateUser(w http.ResponseWriter, r *http.Request) {

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
