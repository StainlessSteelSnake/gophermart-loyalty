package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

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

}

func (h *Handler) getWithdrawals(w http.ResponseWriter, r *http.Request) {

}
