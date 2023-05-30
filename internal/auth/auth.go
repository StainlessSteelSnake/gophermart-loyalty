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
	users          map[string]user
	userController UserAdderGetter
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

func NewAuth(userController UserAdderGetter) (Authenticator, error) {
	if userController == nil {
		return nil, errors.New("не задана функция создания пользователя в БД")
	}

	a := authentication{userController: userController, users: make(map[string]user)}
	return &a, nil
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
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

func (a *authentication) createToken(login, passwordHash string) (string, error) {
	loginHash, err := getHash(login)
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
		login:        login,
		passwordHash: passwordHash,
		tokenSign:    tokenSign,
		tokenFull:    token,
	}

	a.users[loginHash] = user
	return token, nil
}

func (a *authentication) Register(login, password string) (string, error) {
	passwordHash, err := getHash(password)
	if err != nil {
		return "", err
	}

	err = a.userController.AddUser(login, passwordHash)
	if err != nil {
		return "", err
	}

	token, err := a.createToken(login, passwordHash)
	if err != nil {
		return "", err
	}

	return token, err
}

func (a *authentication) Login(login, password string) (string, error) {
	passwordHash, err := a.checkPassword(login, password)
	if err != nil {
		return "", err
	}

	token, err := a.createToken(login, passwordHash)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (a *authentication) checkPassword(login, password string) (string, error) {
	loginHash, err := getHash(login)
	if err != nil {
		return "", err
	}

	savedPasswordHash := ""
	userData, userFound := a.users[loginHash]
	if userFound {
		savedPasswordHash = userData.passwordHash
	}

	if !userFound {
		savedPasswordHash, err = a.userController.GetUserPassword(login)
	}
	if err != nil {
		return "", err
	}

	passwordHash, err := getHash(password)
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
	userData, userFound := a.users[loginHash]
	if !userFound {
		return "", errors.New("указанный пользователь не авторизован")
	}

	if userData.tokenFull != t {
		return "", errors.New("подпись токена авторизации пользователя не соответствует сохранённой")
	}

	return userData.login, nil
}
