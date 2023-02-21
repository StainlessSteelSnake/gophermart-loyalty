package handlers

import (
	"encoding/json"
	"errors"
	"github.com/StainlessSteelSnake/gophermart-loyalty/internal/database"
	"log"
	"net/http"
)

type UserRequestBody struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (h *Handler) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.currentUserLogin = ""

		token := r.Header.Get("Authorization")
		if token == "" || h.auth == nil {
			next.ServeHTTP(w, r)
			return
		}

		var err error
		h.currentUserLogin, err = h.auth.Authenticate(token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		w.Header().Set("Authorization", token)

		next.ServeHTTP(w, r)
	})
}

func (h *Handler) registerUser(w http.ResponseWriter, r *http.Request) {
	request, err := decodeRequest(r)
	if err != nil {
		log.Println("Неверный формат данных в запросе регистрации пользователя:", err)
		http.Error(w, "неверный формат данных в запросе регистрации пользователя: "+err.Error(), http.StatusBadRequest)
		return
	}

	requestBody := UserRequestBody{}
	err = json.Unmarshal(request, &requestBody)
	if err != nil {
		log.Println("Неверный формат данных в запросе регистрации пользователя:", err)
		http.Error(w, "неверный формат данных в запросе регистрации пользователя: "+err.Error(), http.StatusBadRequest)
		return
	}
	log.Println("Переданные данные для регистрации пользователя:", requestBody)

	token, err := h.auth.Register(requestBody.Login, requestBody.Password)
	if err != nil && errors.Is(err, database.DBError{Entity: requestBody.Login, Duplicate: true, Err: nil}) {
		log.Println("Логин уже занят:", err)
		http.Error(w, "логин уже занят: "+err.Error(), http.StatusConflict)
		return
	}

	if err != nil {
		log.Println("Ошибка в сервисе регистрации пользователя:", err)
		http.Error(w, "ошибка в сервисе регистрации пользователя: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", token)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) loginUser(w http.ResponseWriter, r *http.Request) {
	request, err := decodeRequest(r)
	if err != nil {
		log.Println("Неверный формат данных в запросе авторизации пользователя:", err)
		http.Error(w, "неверный формат данных в запросе авторизации пользователя: "+err.Error(), http.StatusBadRequest)
		return
	}

	requestBody := UserRequestBody{}
	err = json.Unmarshal(request, &requestBody)
	if err != nil {
		log.Println("Неверный формат данных в запросе авторизации пользователя:", err)
		http.Error(w, "неверный формат данных в запросе авторизации пользователя: "+err.Error(), http.StatusBadRequest)
		return
	}
	log.Println("Переданные данные для авторизации пользователя:", requestBody)

	token, err := h.auth.Login(requestBody.Login, requestBody.Password)
	if err != nil && errors.Is(err, database.DBError{Entity: requestBody.Login, Duplicate: true, Err: nil}) {
		log.Println("Неверная пара логин/пароль:", err)
		http.Error(w, "неверная пара логин/пароль: "+err.Error(), http.StatusUnauthorized)
		return
	}

	if err != nil {
		log.Println("Ошибка в сервисе авторизации пользователя:", err)
		http.Error(w, "ошибка в сервисе авторизации пользователя: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", token)
	w.WriteHeader(http.StatusOK)
}
