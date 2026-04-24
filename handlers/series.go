package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"proyecto1-STW-backend/models"
)

type SeriesHandler struct {
	DB *sql.DB
}

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, msg string) {
	respondJSON(w, status, map[string]string{"error": msg})
}

func getIDFromPath(path string) (int, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		return 0, fmt.Errorf("no id in path")
	}
	return strconv.Atoi(parts[1])
}

func (h *SeriesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	switch {
	case len(parts) == 1 && r.Method == http.MethodGet:
		h.GetAll(w, r)
	case len(parts) == 1 && r.Method == http.MethodPost:
		h.Create(w, r)
	case len(parts) == 2 && r.Method == http.MethodGet:
		h.GetOne(w, r)
	case len(parts) == 2 && r.Method == http.MethodPut:
		h.Update(w, r)
	case len(parts) == 2 && r.Method == http.MethodDelete:
		h.Delete(w, r)
	case len(parts) == 3 && parts[2] == "rating" && r.Method == http.MethodGet:
		h.GetRating(w, r)
	case len(parts) == 3 && parts[2] == "rating" && r.Method == http.MethodPost:
		h.SetRating(w, r)
	default:
		respondError(w, http.StatusNotFound, "not found")
	}
}

func (h *SeriesHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	search := q.Get("q")
	sort := q.Get("sort")
	order := q.Get("order")
	pageStr := q.Get("page")
	limitStr := q.Get("limit")

	allowed := map[string]bool{"name": true, "current_episode": true, "total_episodes": true, "id": true}
	if !allowed[sort] {
		sort = "id"
	}
	if order != "desc" {
		order = "asc"
	}

	page, limit := 1, 20
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 10000 {
		limit = l
	}
	offset := (page - 1) * limit

	query := fmt.Sprintf(`
		SELECT s.id, s.name, s.current_episode, s.total_episodes, s.image_url,
		       IFNULL(r.rating, 0)
		FROM series s
		LEFT JOIN ratings r ON s.id = r.series_id
		WHERE s.name LIKE ?
		ORDER BY s.%s %s
		LIMIT ? OFFSET ?
	`, sort, order)

	rows, err := h.DB.Query(query, "%"+search+"%", limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "database error")
		return
	}
	defer rows.Close()

	data := []models.SeriesWithRating{}
	for rows.Next() {
		var s models.SeriesWithRating
		rows.Scan(&s.ID, &s.Name, &s.CurrentEpisode, &s.TotalEpisodes, &s.ImageURL, &s.Rating)
		data = append(data, s)
	}

	var total int
	h.DB.QueryRow("SELECT COUNT(*) FROM series WHERE name LIKE ?", "%"+search+"%").Scan(&total)

	respondJSON(w, http.StatusOK, map[string]any{
		"data":  data,
		"page":  page,
		"limit": limit,
		"total": total,
	})
}

func (h *SeriesHandler) GetOne(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromPath(r.URL.Path)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var s models.Series
	err = h.DB.QueryRow(
		"SELECT id, name, current_episode, total_episodes, image_url FROM series WHERE id = ?", id,
	).Scan(&s.ID, &s.Name, &s.CurrentEpisode, &s.TotalEpisodes, &s.ImageURL)

	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "series not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "database error")
		return
	}
	respondJSON(w, http.StatusOK, s)
}

func (h *SeriesHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body models.Series
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if strings.TrimSpace(body.Name) == "" {
		respondError(w, http.StatusBadRequest, "name is required")
		return
	}
	if body.TotalEpisodes <= 0 {
		respondError(w, http.StatusBadRequest, "total_episodes must be greater than 0")
		return
	}
	if body.CurrentEpisode < 0 {
		respondError(w, http.StatusBadRequest, "current_episode cannot be negative")
		return
	}
	if body.CurrentEpisode > body.TotalEpisodes {
		respondError(w, http.StatusBadRequest, "current_episode cannot exceed total_episodes")
		return
	}

	res, err := h.DB.Exec(
		"INSERT INTO series (name, current_episode, total_episodes, image_url) VALUES (?, ?, ?, ?)",
		strings.TrimSpace(body.Name), body.CurrentEpisode, body.TotalEpisodes, body.ImageURL,
	)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not create series")
		return
	}

	insertID, _ := res.LastInsertId()
	body.ID = int(insertID)
	body.Name = strings.TrimSpace(body.Name)
	respondJSON(w, http.StatusCreated, body)
}

func (h *SeriesHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromPath(r.URL.Path)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var exists int
	h.DB.QueryRow("SELECT id FROM series WHERE id = ?", id).Scan(&exists)
	if exists == 0 {
		respondError(w, http.StatusNotFound, "series not found")
		return
	}

	var body models.Series
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if strings.TrimSpace(body.Name) == "" {
		respondError(w, http.StatusBadRequest, "name is required")
		return
	}
	if body.TotalEpisodes <= 0 {
		respondError(w, http.StatusBadRequest, "total_episodes must be greater than 0")
		return
	}
	if body.CurrentEpisode < 0 {
		respondError(w, http.StatusBadRequest, "current_episode cannot be negative")
		return
	}
	if body.CurrentEpisode > body.TotalEpisodes {
		respondError(w, http.StatusBadRequest, "current_episode cannot exceed total_episodes")
		return
	}

	h.DB.Exec(
		"UPDATE series SET name = ?, current_episode = ?, total_episodes = ?, image_url = ? WHERE id = ?",
		strings.TrimSpace(body.Name), body.CurrentEpisode, body.TotalEpisodes, body.ImageURL, id,
	)

	body.ID = id
	body.Name = strings.TrimSpace(body.Name)
	respondJSON(w, http.StatusOK, body)
}

func (h *SeriesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromPath(r.URL.Path)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var exists int
	h.DB.QueryRow("SELECT id FROM series WHERE id = ?", id).Scan(&exists)
	if exists == 0 {
		respondError(w, http.StatusNotFound, "series not found")
		return
	}

	h.DB.Exec("DELETE FROM ratings WHERE series_id = ?", id)
	h.DB.Exec("DELETE FROM series WHERE id = ?", id)
	w.WriteHeader(http.StatusNoContent)
}

func (h *SeriesHandler) GetRating(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromPath(r.URL.Path)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var rating int
	err = h.DB.QueryRow("SELECT rating FROM ratings WHERE series_id = ?", id).Scan(&rating)
	if err == sql.ErrNoRows {
		respondJSON(w, http.StatusOK, map[string]any{"series_id": id, "rating": 0})
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "database error")
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"series_id": id, "rating": rating})
}

func (h *SeriesHandler) SetRating(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromPath(r.URL.Path)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var body struct {
		Rating int `json:"rating"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if body.Rating < 0 || body.Rating > 10 {
		respondError(w, http.StatusBadRequest, "rating must be between 0 and 10")
		return
	}

	h.DB.Exec("DELETE FROM ratings WHERE series_id = ?", id)
	h.DB.Exec("INSERT INTO ratings (series_id, rating) VALUES (?, ?)", id, body.Rating)
	respondJSON(w, http.StatusOK, map[string]any{"series_id": id, "rating": body.Rating})
}
