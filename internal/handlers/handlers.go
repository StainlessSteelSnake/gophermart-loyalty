package handlers

import (
	"encoding/json"
	"errors"
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

func (h *Handler) addOrder(w http.ResponseWriter, r *http.Request) {
	if h.currentUserLogin == "" {
		log.Println("Пользователь не аутентифицирован")
		http.Error(w, "пользователь не аутентифицирован", http.StatusUnauthorized)
		return
	}

	request, err := decodeRequest(r)
	if err != nil {
		log.Println("Неверный формат данных в запросе добавления заказа:", err)
		http.Error(w, "неверный формат данных в запросе добавления заказа: "+err.Error(), http.StatusBadRequest)
		return
	}

	orderID := string(request)

	var orderError *orders.OrderError
	err = h.orders.AddOrder(h.currentUserLogin, orderID)

	if err != nil && errors.As(err, &orderError) {
		log.Println(err)

		switch {
		case orderError.IncorrectID:
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		case orderError.Duplicate && orderError.User == h.currentUserLogin:
			http.Error(w, err.Error(), http.StatusOK)
		case orderError.Duplicate && orderError.User != h.currentUserLogin:
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	if err != nil && errors.Is(err, orders.NewOrderError(orderID, true, false, h.currentUserLogin, nil)) {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("Новый номер заказа '" + orderID + "' принят в обработку")
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) getOrders(w http.ResponseWriter, r *http.Request) {
	if h.currentUserLogin == "" {
		log.Println("Пользователь не аутентифицирован")
		http.Error(w, "пользователь не аутентифицирован", http.StatusUnauthorized)
		return
	}

	orders, err := h.orders.GetOrders(h.currentUserLogin)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		log.Println("Заказы для пользователя не найдены")
		http.Error(w, "заказы для пользователя не найдены", http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	response, err := json.Marshal(orders)
	if err != nil {
		log.Println("Ошибка при формировании ответа:", err)
		http.Error(w, "ошибка при формировании ответа: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

	_, err = w.Write(response)
	if err != nil {
		log.Println("Ошибка при записи ответа в тело запроса:", err)
	}
}

func (h *Handler) getBalance(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) withdrawPoints(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) getWithdrawals(w http.ResponseWriter, r *http.Request) {

}
