package handler

import (
	"encoding/json"
	"habit-tracker/internal/database"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type HabitHandler struct {
	db database.Querier
}

func NewHabitHandler(db database.Querier) *HabitHandler {
	return &HabitHandler{
		db: db,
	}
}

type createHabitRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *HabitHandler) CreateHabit(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value(UserIDKey).(string)
	if !ok {
		http.Error(w, "Не удалось получить ID пользователя", http.StatusInternalServerError)
		return
	}

	var req createHabitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Название привычки не может быть пустым", http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "Неверный формат ID пользователя", http.StatusInternalServerError)
		return
	}

	params := database.CreateHabitParams{
		UserID:      pgtype.UUID{Bytes: userID, Valid: true},
		Name:        req.Name,
		Description: pgtype.Text{String: req.Description, Valid: true},
	}

	habit, err := h.db.CreateHabit(r.Context(), params)
	if err != nil {
		log.Printf("Ошибка создания привычки: %v", err)
		http.Error(w, "Не удалось создать привычку", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(habit)
}

func (h *HabitHandler) ListHabits(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value(UserIDKey).(string)
	if !ok {
		http.Error(w, "Не удалось получить ID пользователя", http.StatusInternalServerError)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "Неверный формат ID пользователя", http.StatusInternalServerError)
		return
	}

	habits, err := h.db.ListHabitsByUserID(r.Context(), pgtype.UUID{Bytes: userID, Valid: true})
	if err != nil {
		log.Printf("Ошибка получения списка привычек: %v", err)
		http.Error(w, "Не удалось получить список привычек", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(habits)
}

func (h *HabitHandler) UpdateHabit(w http.ResponseWriter, r *http.Request) {

	userIDStr, ok := r.Context().Value(UserIDKey).(string)
	if !ok {
		http.Error(w, "Не удалось получить ID пользователя", http.StatusInternalServerError)
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "Неверный формат ID пользователя", http.StatusInternalServerError)
		return
	}

	habitIDStr := chi.URLParam(r, "habitID")
	habitID, err := uuid.Parse(habitIDStr)
	if err != nil {
		http.Error(w, "Неверный формат ID привычки", http.StatusBadRequest)
		return
	}

	var req createHabitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "Название привычки не может быть пустым", http.StatusBadRequest)
		return
	}

	params := database.UpdateHabitParams{
		ID:          pgtype.UUID{Bytes: habitID, Valid: true},
		UserID:      pgtype.UUID{Bytes: userID, Valid: true},
		Name:        req.Name,
		Description: pgtype.Text{String: req.Description, Valid: true},
	}

	updatedHabit, err := h.db.UpdateHabit(r.Context(), params)
	if err != nil {
		log.Printf("Ошибка обновления привычки: %v", err)
		http.Error(w, "Не удалось обновить привычку", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedHabit)
}

func (h *HabitHandler) CheckInHabit(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value(UserIDKey).(string)
	if !ok {
		http.Error(w, "Не удалось получить ID пользователя", http.StatusInternalServerError)
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "Неверный формат ID пользователя", http.StatusInternalServerError)
		return
	}

	habitIDStr := chi.URLParam(r, "habitID")
	habitID, err := uuid.Parse(habitIDStr)
	if err != nil {
		http.Error(w, "Неверный формат ID привычки", http.StatusBadRequest)
		return
	}

	paramsGet := database.GetHabitByIDParams{
		ID:     pgtype.UUID{Bytes: habitID, Valid: true},
		UserID: pgtype.UUID{Bytes: userID, Valid: true},
	}
	_, err = h.db.GetHabitByID(r.Context(), paramsGet)
	if err != nil {

		http.Error(w, "Привычка не найдена или не принадлежит вам", http.StatusNotFound)
		return
	}

	paramsCreate := database.CreateCheckInParams{
		HabitID:     pgtype.UUID{Bytes: habitID, Valid: true},
		CheckInDate: pgtype.Date{Time: time.Now(), Valid: true}, // Используем сегодняшнюю дату
	}

	checkIn, err := h.db.CreateCheckIn(r.Context(), paramsCreate)
	if err != nil {
		log.Printf("Ошибка создания отметки: %v", err)
		http.Error(w, "Не удалось создать отметку (возможно, вы уже отмечались сегодня)", http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(checkIn)
}
