package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"go-http/models"
	"go-http/store"
)

type Handler struct {
	store *store.Store
}

func New(s *store.Store) *Handler {
	return &Handler{store: s}
}

type errorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type successResponse struct {
	Data interface{} `json:"data"`
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{
		Error:   http.StatusText(status),
		Code:    status,
		Message: msg,
	})
}


func (h *Handler) ListWinners(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	winners := h.store.GetAll()

	if idStr := q.Get("id"); idStr != "" {
		id, err := strconv.Atoi(idStr)
		if err != nil || id <= 0 {
			writeError(w, http.StatusBadRequest, "id must be a positive integer")
			return
		}
		winner, ok := h.store.GetByID(id)
		if !ok {
			writeError(w, http.StatusNotFound, "winner not found")
			return
		}
		writeJSON(w, http.StatusOK, successResponse{Data: winner})
		return
	}

	if player := q.Get("player"); player != "" {
		var f []models.Winner
		for _, w := range winners {
			if strings.Contains(strings.ToLower(w.Player), strings.ToLower(player)) {
				f = append(f, w)
			}
		}
		winners = f
	}
	if nat := q.Get("nationality"); nat != "" {
		var f []models.Winner
		for _, w := range winners {
			if strings.EqualFold(w.Nationality, nat) {
				f = append(f, w)
			}
		}
		winners = f
	}
	if club := q.Get("club"); club != "" {
		var f []models.Winner
		for _, w := range winners {
			if strings.Contains(strings.ToLower(w.Club), strings.ToLower(club)) {
				f = append(f, w)
			}
		}
		winners = f
	}
	if yearStr := q.Get("year"); yearStr != "" {
		year, err := strconv.Atoi(yearStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "year must be an integer")
			return
		}
		var f []models.Winner
		for _, w := range winners {
			if w.Year == year {
				f = append(f, w)
			}
		}
		winners = f
	}
	if pos := q.Get("position"); pos != "" {
		var f []models.Winner
		for _, w := range winners {
			if strings.EqualFold(w.Position, pos) {
				f = append(f, w)
			}
		}
		winners = f
	}
	if minStr := q.Get("min_goals"); minStr != "" {
		min, err := strconv.Atoi(minStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "min_goals must be an integer")
			return
		}
		var f []models.Winner
		for _, w := range winners {
			if w.GoalsThatSeason >= min {
				f = append(f, w)
			}
		}
		winners = f
	}

	if winners == nil {
		winners = []models.Winner{}
	}
	writeJSON(w, http.StatusOK, successResponse{Data: winners})
}

func (h *Handler) GetWinner(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	winner, found := h.store.GetByID(id)
	if !found {
		writeError(w, http.StatusNotFound, "winner not found")
		return
	}
	writeJSON(w, http.StatusOK, successResponse{Data: winner})
}

func (h *Handler) CreateWinner(w http.ResponseWriter, r *http.Request) {
	var winner models.Winner
	if err := json.NewDecoder(r.Body).Decode(&winner); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if err := validate(winner); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	created, err := h.store.Add(winner)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not persist winner")
		return
	}
	writeJSON(w, http.StatusCreated, successResponse{Data: created})
}

func (h *Handler) ReplaceWinner(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var winner models.Winner
	if err := json.NewDecoder(r.Body).Decode(&winner); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if err := validate(winner); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	updated, found, err := h.store.Replace(id, winner)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not persist winner")
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "winner not found")
		return
	}
	writeJSON(w, http.StatusOK, successResponse{Data: updated})
}

func (h *Handler) PatchWinner(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var fields map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&fields); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if len(fields) == 0 {
		writeError(w, http.StatusBadRequest, "body must contain at least one field")
		return
	}
	allowed := map[string]bool{
		"player": true, "nationality": true, "club": true,
		"year": true, "votes": true, "position": true, "goals_that_season": true,
	}
	for key := range fields {
		if !allowed[key] {
			writeError(w, http.StatusBadRequest, "unknown field: "+key)
			return
		}
	}
	patched, found, err := h.store.Patch(id, fields)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "winner not found")
		return
	}
	writeJSON(w, http.StatusOK, successResponse{Data: patched})
}

func (h *Handler) DeleteWinner(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	deleted, err := h.store.Delete(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not persist changes")
		return
	}
	if !deleted {
		writeError(w, http.StatusNotFound, "winner not found")
		return
	}
	writeJSON(w, http.StatusOK, successResponse{Data: map[string]string{"message": "winner deleted successfully"}})
}

func parseID(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "id must be a positive integer")
		return 0, false
	}
	return id, true
}

func validate(w models.Winner) error {
	if strings.TrimSpace(w.Player) == "" {
		return valErr("player is required")
	}
	if strings.TrimSpace(w.Nationality) == "" {
		return valErr("nationality is required")
	}
	if strings.TrimSpace(w.Club) == "" {
		return valErr("club is required")
	}
	if w.Year < 1956 || w.Year > 2100 {
		return valErr("year must be between 1956 and 2100 (first Balón de Oro was 1956)")
	}
	if w.Votes < 0 {
		return valErr("votes must be a non-negative number")
	}
	valid := map[string]bool{"Forward": true, "Midfielder": true, "Defender": true, "Goalkeeper": true}
	if !valid[w.Position] {
		return valErr("position must be one of: Forward, Midfielder, Defender, Goalkeeper")
	}
	if w.GoalsThatSeason < 0 {
		return valErr("goals_that_season must be non-negative")
	}
	return nil
}

type valError struct{ msg string }

func (e valError) Error() string { return e.msg }
func valErr(msg string) error    { return valError{msg} }
