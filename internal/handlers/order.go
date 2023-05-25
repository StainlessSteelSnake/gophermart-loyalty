package handlers

import (
	"encoding/json"
	"errors"
	"github.com/StainlessSteelSnake/gophermart-loyalty/internal/orders"
	"log"
	"net/http"
)

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
		log.Println("!", err)

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
