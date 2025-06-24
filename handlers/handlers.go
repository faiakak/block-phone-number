package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// BlockedPhone represents a blocked number record
type BlockedPhone struct {
	ID          int       `json:"id"`
	PhoneNumber string    `json:"phone_number"`
	Reason      string    `json:"reason"`
	BlockedDate time.Time `json:"blocked_date"`
	BlockedBy   string    `json:"blocked_by"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Request structs
type BlockPhoneRequest struct {
	PhoneNumber string `json:"phone_number"`
	Reason      string `json:"reason"`
	BlockedBy   string `json:"blocked_by"`
}

type PhoneCheckRequest struct {
	PhoneNumber string `json:"phone_number"`
}

type PhoneCheckResponse struct {
	IsBlocked   bool   `json:"is_blocked"`
	PhoneNumber string `json:"phone_number"`
	Reason      string `json:"reason,omitempty"`
	BlockedBy   string `json:"blocked_by,omitempty"`
	BlockedDate string `json:"blocked_date,omitempty"`
}

// Global database reference (set from main)
var DB *sql.DB

func SetDB(db *sql.DB) {
	DB = db
}

// --- Handlers ---

func GetBlockedPhones(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query := `
		SELECT id, phone_number, reason, blocked_date, blocked_by, is_active, created_at, updated_at 
		FROM emaginenet_blocked_numbers 
		WHERE is_active = true 
		ORDER BY blocked_date DESC
	`

	rows, err := DB.Query(query)
	if err != nil {
		log.Printf("Error querying blocked phones: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var phones []BlockedPhone
	for rows.Next() {
		var phone BlockedPhone
		err := rows.Scan(
			&phone.ID, &phone.PhoneNumber, &phone.Reason, &phone.BlockedDate,
			&phone.BlockedBy, &phone.IsActive, &phone.CreatedAt, &phone.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		phones = append(phones, phone)
	}

	json.NewEncoder(w).Encode(phones)
}

func AddBlockedPhone(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req BlockPhoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.PhoneNumber) == "" {
		http.Error(w, "Phone number is required", http.StatusBadRequest)
		return
	}

	if !validatePhoneNumber(req.PhoneNumber) {
		http.Error(w, "Invalid phone number format", http.StatusBadRequest)
		return
	}

	normalizedPhone := normalizePhoneNumber(req.PhoneNumber)

	if strings.TrimSpace(req.Reason) == "" {
		req.Reason = "No reason provided"
	}
	if strings.TrimSpace(req.BlockedBy) == "" {
		req.BlockedBy = "System"
	}

	// First, check if phone number exists and its current status
	var existingID int
	var existingIsActive bool
	checkQuery := `SELECT id, is_active FROM emaginenet_blocked_numbers WHERE phone_number = $1 ORDER BY blocked_date DESC LIMIT 1`
	err := DB.QueryRow(checkQuery, normalizedPhone).Scan(&existingID, &existingIsActive)

	var phone BlockedPhone

	if err == sql.ErrNoRows {
		// Phone number doesn't exist, create new record
		query := `
			INSERT INTO emaginenet_blocked_numbers (phone_number, reason, blocked_by) 
			VALUES ($1, $2, $3) 
			RETURNING id, phone_number, reason, blocked_date, blocked_by, is_active, created_at, updated_at
		`
		err = DB.QueryRow(query, normalizedPhone, req.Reason, req.BlockedBy).Scan(
			&phone.ID, &phone.PhoneNumber, &phone.Reason, &phone.BlockedDate,
			&phone.BlockedBy, &phone.IsActive, &phone.CreatedAt, &phone.UpdatedAt,
		)
	} else if err != nil {
		// Database error
		log.Printf("Error checking existing blocked phone: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	} else if existingIsActive {
		// Phone number is already actively blocked
		http.Error(w, "Phone number is already blocked", http.StatusConflict)
		return
	} else {
		// Phone number exists but is inactive, create new blocking record (re-block)
		query := `
			INSERT INTO emaginenet_blocked_numbers (phone_number, reason, blocked_by) 
			VALUES ($1, $2, $3) 
			RETURNING id, phone_number, reason, blocked_date, blocked_by, is_active, created_at, updated_at
		`
		err = DB.QueryRow(query, normalizedPhone, req.Reason, req.BlockedBy).Scan(
			&phone.ID, &phone.PhoneNumber, &phone.Reason, &phone.BlockedDate,
			&phone.BlockedBy, &phone.IsActive, &phone.CreatedAt, &phone.UpdatedAt,
		)
	}

	if err != nil {
		log.Printf("Error inserting blocked phone: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(phone)
}

func RemoveBlockedPhone(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	query := `UPDATE emaginenet_blocked_numbers SET is_active = false WHERE id = $1 AND is_active = true`
	result, err := DB.Exec(query, id)
	if err != nil {
		log.Printf("Error removing blocked phone: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Phone number not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Phone number removed from blocked list"})
}

func CheckPhone(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req PhoneCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.PhoneNumber) == "" {
		http.Error(w, "Phone number is required", http.StatusBadRequest)
		return
	}

	normalizedPhone := normalizePhoneNumber(req.PhoneNumber)

	query := `
		SELECT phone_number, reason, blocked_by, blocked_date 
		FROM emaginenet_blocked_numbers 
		WHERE phone_number = $1 AND is_active = true
		ORDER BY blocked_date DESC
		LIMIT 1
	`

	var phone BlockedPhone
	err := DB.QueryRow(query, normalizedPhone).Scan(
		&phone.PhoneNumber, &phone.Reason, &phone.BlockedBy, &phone.BlockedDate,
	)

	response := PhoneCheckResponse{
		PhoneNumber: normalizedPhone,
	}

	if err == sql.ErrNoRows {
		response.IsBlocked = false
	} else if err != nil {
		log.Printf("Error checking phone: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	} else {
		response.IsBlocked = true
		response.Reason = phone.Reason
		response.BlockedBy = phone.BlockedBy
		response.BlockedDate = phone.BlockedDate.Format("2006-01-02 15:04:05")
	}

	json.NewEncoder(w).Encode(response)
}

// GetPhoneHistory returns the complete block/unblock history for a specific phone number
func GetPhoneHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	phoneNumber := vars["phoneNumber"]
	
	if phoneNumber == "" {
		http.Error(w, "Phone number is required", http.StatusBadRequest)
		return
	}

	normalizedPhone := normalizePhoneNumber(phoneNumber)

	query := `
		SELECT id, phone_number, reason, blocked_date, blocked_by, is_active, created_at, updated_at 
		FROM emaginenet_blocked_numbers 
		WHERE phone_number = $1 
		ORDER BY blocked_date DESC
	`

	rows, err := DB.Query(query, normalizedPhone)
	if err != nil {
		log.Printf("Error querying phone history: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var history []BlockedPhone
	for rows.Next() {
		var record BlockedPhone
		err := rows.Scan(
			&record.ID, &record.PhoneNumber, &record.Reason, &record.BlockedDate,
			&record.BlockedBy, &record.IsActive, &record.CreatedAt, &record.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning history row: %v", err)
			continue
		}
		history = append(history, record)
	}

	if len(history) == 0 {
		// Return empty array instead of null for better frontend handling
		json.NewEncoder(w).Encode([]BlockedPhone{})
		return
	}

	json.NewEncoder(w).Encode(history)
}

// --- Utilities ---

func normalizePhoneNumber(phone string) string {
	re := regexp.MustCompile(`[^\d]`)
	cleaned := re.ReplaceAllString(phone, "")

	if len(cleaned) == 10 {
		return fmt.Sprintf("(%s) %s-%s", cleaned[:3], cleaned[3:6], cleaned[6:])
	}
	if len(cleaned) == 11 && cleaned[0] == '1' {
		return fmt.Sprintf("1-(%s) %s-%s", cleaned[1:4], cleaned[4:7], cleaned[7:])
	}
	return phone
}

func validatePhoneNumber(phone string) bool {
	re := regexp.MustCompile(`[^\d]`)
	cleaned := re.ReplaceAllString(phone, "")
	return len(cleaned) == 10 || (len(cleaned) == 11 && cleaned[0] == '1')
}

// --- CORS Middleware ---

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
