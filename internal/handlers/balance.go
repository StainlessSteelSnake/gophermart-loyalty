package handlers

import (
	"encoding/json"
	"errors"
	"github.com/StainlessSteelSnake/gophermart-loyalty/internal/orders"
	"log"
	"net/http"
)

type WithdrawRequestBody struct {
	OrderID string  `json:"order"`
	Amount  float32 `json:"sum"`
}

func (h *Handler) getBalance(w http.ResponseWriter, r *http.Request) {
	if h.currentUserLogin == "" {
		log.Println("Пользователь не аутентифицирован")
		http.Error(w, "пользователь не аутентифицирован", http.StatusUnauthorized)
		return
	}

	account, err := h.orders.GetUserAccount(h.currentUserLogin)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	response, err := json.Marshal(account)
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

func (h *Handler) withdrawPoints(w http.ResponseWriter, r *http.Request) {
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

	requestBody := WithdrawRequestBody{}
	err = json.Unmarshal(request, &requestBody)
	if err != nil {
		log.Println("Неверный формат данных в запросе на списание средств:", err)
		http.Error(w, "неверный формат данных в запросе на списание средств: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Println("Переданные данные для списания средств:", requestBody)

	var orderError *orders.OrderError
	err = h.orders.WithdrawForOrder(h.currentUserLogin, requestBody.OrderID, requestBody.Amount)

	if err != nil && errors.As(err, &orderError) {
		log.Println("Ошибка при обработке запроса на списание средств: " + err.Error())

		switch {
		case orderError.IncorrectID:
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		case orderError.InsufficientFunds:
			http.Error(w, err.Error(), http.StatusPaymentRequired)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	if err != nil {
		log.Println("Ошибка при обработке запроса на списание средств: " + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) getWithdrawals(w http.ResponseWriter, r *http.Request) {

}
