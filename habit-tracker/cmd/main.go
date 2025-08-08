package main

import (
	"context"
	"fmt"
	"habit-tracker/internal/database"
	"habit-tracker/internal/handler"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Ошибка загрузки .env файла")
	}

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("Переменная окружения DATABASE_URL не установлена")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("Переменная окружения JWT_SECRET не установлена")
	}

	dbpool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}
	defer dbpool.Close()

	if err := dbpool.Ping(context.Background()); err != nil {
		log.Fatalf("Не удалось пинговать базу данных: %v", err)
	}
	fmt.Println("Успешное подключение к базе данных!")

	dbQueries := database.New(dbpool)
	userHandler := handler.NewUserHandler(dbQueries, jwtSecret)
	habitHandler := handler.NewHabitHandler(dbQueries)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Сервер работает!"))
	})
	r.Post("/register", userHandler.Register)
	r.Post("/login", userHandler.Login)

	r.Group(func(r chi.Router) {
		r.Use(userHandler.AuthMiddleware)

		r.Get("/api/me", func(w http.ResponseWriter, r *http.Request) {
			userID := r.Context().Value(handler.UserIDKey).(string)
			w.Write([]byte("Привет! Твой ID пользователя: " + userID))
		})

		r.Post("/api/habits", habitHandler.CreateHabit)
		r.Get("/api/habits", habitHandler.ListHabits)
		r.Put("/api/habits/{habitID}", habitHandler.UpdateHabit)
		r.Post("/api/habits/{habitID}/checkin", habitHandler.CheckInHabit)
	})

	log.Println("Успешный запуск сервера на порту http://localhost:8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Сервер не запущен: %v", err)
	}
}
