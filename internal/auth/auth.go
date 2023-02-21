package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

const secretKey = "TheSecretKey"

type user struct {
	login        string
	passwordHash string
	tokenSign    string
	tokenFull    string
}

type authentication struct {
	userController UserAdderGetter
	users          map[string]user
}

type UserAdderGetter interface {
	AddUser(string, string) error
	GetUserPassword(string) (string, error)
}

type Authenticator interface {
	Authenticate(string) (string, error)
	Register(string, string) (string, error)
	Login(string, string) (string, error)
}

func NewAuth(u UserAdderGetter) Authenticator {
	a := authentication{userController: u, users: make(map[string]user)}
	return &a
}

func getHash(s string) (string, error) {
	hasher := sha256.New()

	_, err := hasher.Write([]byte(s))
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func getSign(s string) (string, error) {
	h := hmac.New(sha256.New, []byte(secretKey))

	_, err := h.Write([]byte(s))
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func getRandom(size int) (string, error) {
	// генерируем случайную последовательность байт
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

func (a *authentication) createToken(l, ph string) (string, error) {
	loginHash, err := getHash(l)
	if err != nil {
		return "", err
	}

	random, err := getRandom(6)
	if err != nil {
		return "", err
	}

	token := loginHash + ":" + random

	tokenSign, err := getSign(token)
	if err != nil {
		return "", err
	}

	token = token + tokenSign

	user := user{
		login:        l,
		passwordHash: ph,
		tokenSign:    tokenSign,
		tokenFull:    token,
	}

	a.users[loginHash] = user
	return token, nil
}

func (a *authentication) Register(l, p string) (string, error) {
	if a.userController == nil {
		return "", errors.New("не задана функция создания пользователя в БД")
	}

	passwordHash, err := getHash(p)
	if err != nil {
		return "", err
	}

	err = a.userController.AddUser(l, passwordHash)
	if err != nil {
		return "", err
	}

	token, err := a.createToken(l, passwordHash)
	if err != nil {
		return "", err
	}

	return token, err
}

func (a *authentication) Login(l, p string) (string, error) {
	passwordHash, err := a.checkPassword(l, p)
	if err != nil {
		return "", err
	}

	token, err := a.createToken(l, passwordHash)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (a *authentication) checkPassword(l, p string) (string, error) {
	if a.userController == nil {
		return "", errors.New("не задана функция получения из БД хэша сохранённого пароля")
	}

	savedPasswordHash, err := a.userController.GetUserPassword(l)
	if err != nil {
		return "", err
	}

	passwordHash, err := getHash(p)
	if err != nil {
		return "", err
	}

	if savedPasswordHash != passwordHash {
		return "", errors.New("переданный и сохранённый пароли не совпадают")
	}

	return passwordHash, nil
}

func (a *authentication) Authenticate(t string) (string, error) {
	tokenParts := strings.Split(t, ":")
	if len(tokenParts) != 2 {
		return "", errors.New("токен авторизации передан в неправильном формате")
	}

	loginHash := tokenParts[0]
	user, ok := a.users[loginHash]
	if !ok {
		return "", errors.New("указанный пользователь не авторизован")
	}

	if user.tokenFull != t {
		return "", errors.New("подпись токена авторизации пользователя не соответствует сохранённой")
	}

	return user.login, nil
}
