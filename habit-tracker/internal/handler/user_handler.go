package handler

import (
	"encoding/json"
	"errors"
	"habit-tracker/internal/database"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	db        database.Querier
	jwtSecret []byte
}

func NewUserHandler(db database.Querier, jwtSecret string) *UserHandler {
	return &UserHandler{
		db:        db,
		jwtSecret: []byte(jwtSecret),
	}
}

type authRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string `json:"token"`
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Ошибка при чтении запроса", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, "Юзернейм и пароль не могут быть пустыми", http.StatusBadRequest)
		return
	}

	_, err = h.db.GetUserByUsername(r.Context(), req.Username)
	if err == nil {
		http.Error(w, "Пользователь с таким юзернеймом уже существует", http.StatusConflict)
		return
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		log.Printf("Внутрення ошибка сервера при проверке пользовател: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Ошибка при хешировании пароля: %v", err)
		http.Error(w, "Ошибка при хешировании пароля", http.StatusInternalServerError)
		return
	}

	params := database.CreateUserParams{
		Username:     req.Username,
		PasswordHash: string(hashedPassword),
	}

	_, err = h.db.CreateUser(r.Context(), params)
	if err != nil {
		log.Printf("Ошибка при создании пользователя в БД: %v", err)
		http.Error(w, "Ошибка создания пользователя", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Пользователь успешно создан"))
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {

	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Ошибка при чтении тела запроса", http.StatusBadRequest)
		return
	}

	user, err := h.db.GetUserByUsername(r.Context(), req.Username)
	if err != nil {
		http.Error(w, "Неверное имя пользователя или пароль", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		http.Error(w, "Неверное имя пользователя или пароль", http.StatusUnauthorized)
		return
	}

	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(h.jwtSecret)
	if err != nil {
		log.Printf("Ошибка при создании токена: %v", err)
		http.Error(w, "Не удалось создать токен", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(authResponse{Token: tokenString})

}
