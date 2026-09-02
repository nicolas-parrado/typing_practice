package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/nparrado/typing-exercises/backend/db"
	"github.com/nparrado/typing-exercises/backend/models"
)

// JSON helper response
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("HTML 500: Internal Server Error"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// GetProfiles handles GET /api/profiles
func GetProfiles(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query("SELECT id, name, xp, level, streak, last_active, theme, sound_enabled, switch_type, created_at FROM profiles ORDER BY name ASC")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	profiles := []models.Profile{}
	for rows.Next() {
		var p models.Profile
		var createdAt string
		var soundEnabledVal int
		err := rows.Scan(&p.ID, &p.Name, &p.XP, &p.Level, &p.Streak, &p.LastActive, &p.Theme, &soundEnabledVal, &p.SwitchType, &createdAt)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		p.SoundEnabled = soundEnabledVal == 1
		p.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", strings.Split(createdAt, ".")[0])
		profiles = append(profiles, p)
	}

	respondWithJSON(w, http.StatusOK, profiles)
}

// CreateProfile handles POST /api/profiles
func CreateProfile(w http.ResponseWriter, r *http.Request) {
	var p models.Profile
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	p.Name = strings.TrimSpace(p.Name)
	if p.Name == "" {
		respondWithError(w, http.StatusBadRequest, "Profile name cannot be empty")
		return
	}

	res, err := db.DB.Exec("INSERT INTO profiles (name, xp, level, streak, last_active, theme, sound_enabled, switch_type) VALUES (?, 0, 1, 0, '', 'glass', 1, 'blue')", p.Name)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			respondWithError(w, http.StatusConflict, "Profile name already exists")
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	p.ID = int(id)
	p.XP = 0
	p.Level = 1
	p.Streak = 0
	p.LastActive = ""
	p.Theme = "glass"
	p.SoundEnabled = true
	p.SwitchType = "blue"
	p.CreatedAt = time.Now()

	// Automatically unlock achievement #1: Bienvenido a la Tropa
	_, _ = db.DB.Exec("INSERT OR IGNORE INTO achievements (profile_id, code) VALUES (?, 'welcome')", p.ID)

	respondWithJSON(w, http.StatusCreated, p)
}

// GetProfile handles GET /api/profiles/{id}
func GetProfile(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid profile ID")
		return
	}

	var p models.Profile
	var createdAt string
	var soundEnabledVal int
	err = db.DB.QueryRow("SELECT id, name, xp, level, streak, last_active, theme, sound_enabled, switch_type, created_at FROM profiles WHERE id = ?", id).
		Scan(&p.ID, &p.Name, &p.XP, &p.Level, &p.Streak, &p.LastActive, &p.Theme, &soundEnabledVal, &p.SwitchType, &createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Profile not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	p.SoundEnabled = soundEnabledVal == 1

	respondWithJSON(w, http.StatusOK, p)
}

// UpdateProfileSettingsPayload represents settings update body
type UpdateProfileSettingsPayload struct {
	Theme        string `json:"theme"`
	SoundEnabled bool   `json:"sound_enabled"`
	SwitchType   string `json:"switch_type"`
}

// UpdateProfileSettings handles PUT /api/profiles/{id}/settings
func UpdateProfileSettings(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid profile ID")
		return
	}

	var payload UpdateProfileSettingsPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	soundVal := 0
	if payload.SoundEnabled {
		soundVal = 1
	}

	_, err = db.DB.Exec("UPDATE profiles SET theme = ?, sound_enabled = ?, switch_type = ? WHERE id = ?",
		payload.Theme, soundVal, payload.SwitchType, id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

// DeleteProfile handles DELETE /api/profiles/{id}
func DeleteProfile(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid profile ID")
		return
	}

	_, err = db.DB.Exec("DELETE FROM profiles WHERE id = ?", id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

// GetExercises handles GET /api/exercises
func GetExercises(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	difficulty := r.URL.Query().Get("difficulty")
	isEnduranceStr := r.URL.Query().Get("is_endurance")
	profileIdStr := r.URL.Query().Get("profile_id")

	// 1. Fetch user scores history if profile_id is provided
	type scoreRecord struct {
		classicWPM float64
		classicAcc float64
		arcadeWPM  float64
		arcadeAcc  float64
		playCount  int
		passed     bool
	}
	scores := make(map[string]*scoreRecord)

	if profileIdStr != "" {
		profileID, err := strconv.Atoi(profileIdStr)
		if err == nil {
			rows, err := db.DB.Query(`SELECT exercise_id, mode, MAX(wpm), MAX(accuracy), COUNT(id) 
				FROM sessions WHERE profile_id = ? GROUP BY exercise_id, mode`, profileID)
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					var exID, mode string
					var wpm, acc float64
					var count int
					if err := rows.Scan(&exID, &mode, &wpm, &acc, &count); err == nil {
						if _, exists := scores[exID]; !exists {
							scores[exID] = &scoreRecord{}
						}
						rec := scores[exID]
						rec.playCount += count
						if mode == "arcade" {
							rec.arcadeWPM = wpm
							rec.arcadeAcc = acc
						} else {
							// classic / lesson
							rec.classicWPM = wpm
							rec.classicAcc = acc
						}
						// Mark passed if WPM >= 25 and Accuracy >= 90%
						if wpm >= 25.0 && acc >= 0.90 {
							rec.passed = true
						}
					}
				}
			}
		}
	}

	// 2. Fetch Exercises, ordered numerically/alphabetically
	query := "SELECT id, title, content, category, difficulty, is_endurance, stage, target_wpm, min_accuracy FROM exercises WHERE 1=1"
	args := []interface{}{}

	if category != "" {
		query += " AND category = ?"
		args = append(args, category)
	}
	if difficulty != "" {
		query += " AND difficulty = ?"
		args = append(args, difficulty)
	}
	if isEnduranceStr != "" {
		isEndurance := 0
		if isEnduranceStr == "true" || isEnduranceStr == "1" {
			isEndurance = 1
		}
		query += " AND is_endurance = ?"
		args = append(args, isEndurance)
	}

	// Sort numerically using the length trick
	query += " ORDER BY length(id) ASC, id ASC"

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	exercises := []models.Exercise{}
	for rows.Next() {
		var e models.Exercise
		var isEndInt int
		err := rows.Scan(&e.ID, &e.Title, &e.Content, &e.Category, &e.Difficulty, &isEndInt, &e.Stage, &e.TargetWPM, &e.MinAccuracy)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		e.IsEndurance = (isEndInt == 1)
		exercises = append(exercises, e)
	}

	// 3. Calculate locks and bind scores adaptively
	previousPassed := true
	for idx := range exercises {
		ex := &exercises[idx]
		exScore := scores[ex.ID]
		if exScore != nil {
			ex.HighScoreWPM = exScore.classicWPM
			ex.HighScoreAccuracy = exScore.classicAcc
			ex.ArcadeWPM = exScore.arcadeWPM
			ex.ArcadeAccuracy = exScore.arcadeAcc
			ex.PlayCount = exScore.playCount
		}

		if idx == 0 {
			ex.Unlocked = true
		} else {
			ex.Unlocked = previousPassed
		}

		hasPassed := false
		if exScore != nil {
			// Check against the lesson's specific adaptive targets
			if (exScore.classicWPM >= ex.TargetWPM && exScore.classicAcc >= ex.MinAccuracy) ||
				(exScore.arcadeWPM >= ex.TargetWPM && exScore.arcadeAcc >= ex.MinAccuracy) {
				hasPassed = true
			}
		}
		previousPassed = hasPassed
	}

	respondWithJSON(w, http.StatusOK, exercises)
}

// SaveSessionPayload represents incoming typing result
type SaveSessionPayload struct {
	ProfileID       int                    `json:"profile_id"`
	ExerciseID      string                 `json:"exercise_id"`
	Mode            string                 `json:"mode"`
	WPM             float64                `json:"wpm"`
	Accuracy        float64                `json:"accuracy"`
	DurationSeconds int                    `json:"duration_seconds"`
	RawDataJSON     string                 `json:"raw_data_json"`
	BackspacesUsed  int                    `json:"backspaces_used"`
	Keys            map[string]KeyProgress `json:"keys"`
}

type KeyProgress struct {
	Attempts int `json:"attempts"`
	Errors   int `json:"errors"`
	Latency  int `json:"latency_ms"`
}

// SaveSession handles POST /api/sessions
func SaveSession(w http.ResponseWriter, r *http.Request) {
	var payload SaveSessionPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	tx, err := db.DB.Begin()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer tx.Rollback()

	// 1. Insert session record
	_, err = tx.Exec(`INSERT INTO sessions (profile_id, exercise_id, mode, wpm, accuracy, duration_seconds, raw_data_json) 
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		payload.ProfileID, payload.ExerciseID, payload.Mode, payload.WPM, payload.Accuracy, payload.DurationSeconds, payload.RawDataJSON)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 2. Manage Failed/Retry states using exercise target or default
	var minAcc float64 = 0.90
	_ = tx.QueryRow("SELECT min_accuracy FROM exercises WHERE id = ?", payload.ExerciseID).Scan(&minAcc)
	if payload.Accuracy < minAcc {
		// Insert or replace in failed attempts
		_, err = tx.Exec(`INSERT OR REPLACE INTO failed_attempts 
			(profile_id, exercise_id, wpm, accuracy, created_at) 
			VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)`,
			payload.ProfileID, payload.ExerciseID, payload.WPM, payload.Accuracy)
	} else if payload.Accuracy >= minAcc+0.05 || payload.Accuracy >= 0.95 {
		// Clean up from retry queue since they mastered it
		_, err = tx.Exec(`DELETE FROM failed_attempts WHERE profile_id = ? AND exercise_id = ?`,
			payload.ProfileID, payload.ExerciseID)
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 3. Update Profile XP & Level and Check Streaks
	var currentXP, currentLevel, currentStreak int
	var lastActive string
	err = tx.QueryRow("SELECT xp, level, streak, last_active FROM profiles WHERE id = ?", payload.ProfileID).
		Scan(&currentXP, &currentLevel, &currentStreak, &lastActive)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Calculate XP gained
	xpGained := int(payload.WPM * payload.Accuracy * 1.5)
	if payload.Mode == "endurance" {
		xpGained *= 2 // Double XP for long endurance tests
	}
	if xpGained < 10 {
		xpGained = 10 // Minimum XP for completing
	}

	newXP := currentXP + xpGained
	newLevel := currentLevel
	for newXP >= newLevel*1000 {
		newXP -= newLevel * 1000
		newLevel++
	}

	// Calculate Streak
	todayStr := time.Now().Format("2006-01-02")
	yesterdayStr := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	newStreak := currentStreak
	if lastActive == "" {
		newStreak = 1
	} else if lastActive == yesterdayStr {
		newStreak++
	} else if lastActive != todayStr {
		// Not active yesterday and not active today, streak broke
		newStreak = 1
	}

	_, err = tx.Exec("UPDATE profiles SET xp = ?, level = ?, streak = ?, last_active = ? WHERE id = ?",
		newXP, newLevel, newStreak, todayStr, payload.ProfileID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 4. Update Key Metrics
	for char, progress := range payload.Keys {
		if char == "" {
			continue
		}
		// Try to insert if not exists, otherwise update
		var attempts, errors, avgLatency int
		err = tx.QueryRow("SELECT attempts, errors, avg_latency_ms FROM key_metrics WHERE profile_id = ? AND char = ?",
			payload.ProfileID, char).Scan(&attempts, &errors, &avgLatency)

		if err == sql.ErrNoRows {
			_, err = tx.Exec(`INSERT INTO key_metrics (profile_id, char, attempts, errors, avg_latency_ms) 
				VALUES (?, ?, ?, ?, ?)`,
				payload.ProfileID, char, progress.Attempts, progress.Errors, progress.Latency)
		} else if err == nil {
			newAttempts := attempts + progress.Attempts
			newErrors := errors + progress.Errors
			// Calculate new weighted latency average
			newLatency := avgLatency
			if newAttempts > 0 && progress.Attempts > 0 {
				newLatency = ((avgLatency * attempts) + (progress.Latency * progress.Attempts)) / newAttempts
			}
			_, err = tx.Exec(`UPDATE key_metrics SET attempts = ?, errors = ?, avg_latency_ms = ? 
				WHERE profile_id = ? AND char = ?`,
				newAttempts, newErrors, newLatency, payload.ProfileID, char)
		}
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	if err := tx.Commit(); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 5. Evaluate and Unlock Achievements
	newAchievements := evaluateAchievements(payload.ProfileID, payload, newStreak)

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"xp_gained":        xpGained,
		"new_xp":           newXP,
		"new_level":        newLevel,
		"new_streak":       newStreak,
		"level_up":         newLevel > currentLevel,
		"new_achievements": newAchievements,
	})
}

// Evaluate medals unlocked during this session
func evaluateAchievements(profileID int, payload SaveSessionPayload, streak int) []string {
	newlyUnlocked := []string{}
	now := time.Now()

	// Get already unlocked achievements
	rows, err := db.DB.Query("SELECT code FROM achievements WHERE profile_id = ?", profileID)
	if err != nil {
		return newlyUnlocked
	}
	defer rows.Close()

	unlockedMap := make(map[string]bool)
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err == nil {
			unlockedMap[code] = true
		}
	}

	// Helper to unlock
	unlock := func(code string) {
		if !unlockedMap[code] {
			_, err = db.DB.Exec("INSERT OR IGNORE INTO achievements (profile_id, code, unlocked_at) VALUES (?, ?, ?)",
				profileID, code, now)
			if err == nil {
				newlyUnlocked = append(newlyUnlocked, code)
				unlockedMap[code] = true
			}
		}
	}

	// Metric checks
	var totalSessions, codeSessions, enduranceSessions, perfectSessions, wpmOver60, wpmOver90 int
	var maxWPM, maxCodeWPM float64

	db.DB.QueryRow("SELECT COUNT(*) FROM sessions WHERE profile_id = ?", profileID).Scan(&totalSessions)
	db.DB.QueryRow(`SELECT COUNT(s.id) FROM sessions s 
		JOIN exercises e ON s.exercise_id = e.id 
		WHERE s.profile_id = ? AND e.category = 'code'`, profileID).Scan(&codeSessions)
	db.DB.QueryRow(`SELECT COUNT(s.id) FROM sessions s 
		JOIN exercises e ON s.exercise_id = e.id 
		WHERE s.profile_id = ? AND e.is_endurance = 1`, profileID).Scan(&enduranceSessions)
	db.DB.QueryRow("SELECT COUNT(*) FROM sessions WHERE profile_id = ? AND accuracy = 1.0", profileID).Scan(&perfectSessions)
	db.DB.QueryRow("SELECT IFNULL(MAX(wpm), 0) FROM sessions WHERE profile_id = ?", profileID).Scan(&maxWPM)
	db.DB.QueryRow(`SELECT IFNULL(MAX(s.wpm), 0) FROM sessions s 
		JOIN exercises e ON s.exercise_id = e.id 
		WHERE s.profile_id = ? AND e.category = 'code'`, profileID).Scan(&maxCodeWPM)
	db.DB.QueryRow("SELECT COUNT(*) FROM sessions WHERE profile_id = ? AND wpm >= 60", profileID).Scan(&wpmOver60)
	db.DB.QueryRow("SELECT COUNT(*) FROM sessions WHERE profile_id = ? AND wpm >= 90", profileID).Scan(&wpmOver90)

	// Get consecutive perfect sessions
	var consecutivePerfect int
	pRows, err := db.DB.Query("SELECT accuracy FROM sessions WHERE profile_id = ? ORDER BY completed_at DESC LIMIT 20", profileID)
	if err == nil {
		for pRows.Next() {
			var acc float64
			if pRows.Scan(&acc) == nil && acc == 1.0 {
				consecutivePerfect++
			} else {
				break
			}
		}
		pRows.Close()
	}

	// 1. Bienvenido a la Tropa (done on profile creation, but check just in case)
	unlock("welcome")

	// 2. Puesta en marcha
	if totalSessions >= 1 {
		unlock("first_step")
	}
	// 3. Calentando motores
	if payload.WPM >= 30 {
		unlock("warm_up")
	}
	// 4. Buen ritmo
	if payload.Accuracy >= 0.95 {
		unlock("good_pace")
	}
	// 5. Políglota Novato
	var hasSpa, hasEng bool
	db.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM sessions s JOIN exercises e ON s.exercise_id = e.id WHERE s.profile_id = ? AND e.category = 'spanish')`, profileID).Scan(&hasSpa)
	db.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM sessions s JOIN exercises e ON s.exercise_id = e.id WHERE s.profile_id = ? AND e.category = 'english')`, profileID).Scan(&hasEng)
	if hasSpa && hasEng {
		unlock("polyglot_novice")
	}
	// 6. Primer Script
	if codeSessions >= 1 {
		unlock("first_script")
	}
	// 7. Racha de Iniciación
	if streak >= 2 {
		unlock("streak_inits")
	}
	// 8. Sin prisas
	if enduranceSessions >= 1 {
		unlock("take_your_time")
	}

	// TIER 2
	// 9. Mecanógrafo Constante
	if streak >= 7 {
		unlock("streak_constant")
	}
	// 10. Desarrollador Junior
	if codeSessions >= 15 {
		unlock("dev_junior")
	}
	// 11. Ojo de Halcón
	if payload.Accuracy == 1.0 && payload.DurationSeconds >= 10 { // Ensure it was a real test
		unlock("hawk_eye")
	}
	// 12. Velocista de Plata
	if maxWPM >= 60 {
		unlock("speedster_silver")
	}
	// 13. Resistencia de Bronce
	if enduranceSessions >= 5 {
		unlock("endurance_bronze")
	}
	// 14. El Retorno
	if payload.Mode == "retry" && payload.Accuracy >= 0.95 {
		unlock("the_return")
	}
	// 15. Maestro de Símbolos
	var symbolCount int
	db.DB.QueryRow(`SELECT COUNT(*) FROM sessions s JOIN exercises e ON s.exercise_id = e.id 
		WHERE s.profile_id = ? AND e.category IN ('symbols', 'numbers')`, profileID).Scan(&symbolCount)
	if symbolCount >= 10 {
		unlock("symbol_master")
	}
	// 16. Tarde de Práctica
	var totalDurationToday int
	db.DB.QueryRow("SELECT SUM(duration_seconds) FROM sessions WHERE profile_id = ? AND DATE(completed_at) = DATE('now')", profileID).Scan(&totalDurationToday)
	if totalDurationToday >= 1800 { // 30 minutes
		unlock("practice_afternoon")
	}
	// 17. Cero Borrados
	if payload.BackspacesUsed == 0 && payload.DurationSeconds > 15 {
		unlock("zero_backspace")
	}

	// TIER 3
	// 18. Mecanógrafo de Oro
	if maxWPM >= 90 {
		unlock("speedster_gold")
	}
	// 19. Racha de Plata
	if streak >= 15 {
		unlock("streak_silver")
	}
	// 20. Desarrollador Senior
	var highAccCode int
	db.DB.QueryRow(`SELECT COUNT(*) FROM sessions s JOIN exercises e ON s.exercise_id = e.id 
		WHERE s.profile_id = ? AND e.category = 'code' AND s.accuracy >= 0.97`, profileID).Scan(&highAccCode)
	if highAccCode >= 50 {
		unlock("dev_senior")
	}
	// 21. Precisión Milimétrica
	var consecutiveHighAcc int
	hRows, err := db.DB.Query("SELECT accuracy, wpm FROM sessions WHERE profile_id = ? ORDER BY completed_at DESC LIMIT 5", profileID)
	if err == nil {
		for hRows.Next() {
			var acc, wpm float64
			if hRows.Scan(&acc, &wpm) == nil && acc >= 0.99 && wpm >= 50 {
				consecutiveHighAcc++
			} else {
				break
			}
		}
		hRows.Close()
	}
	if consecutiveHighAcc >= 5 {
		unlock("precision_perfect")
	}
	// 22. Maratonista de Acero
	var maratonistaAcero bool
	db.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM sessions s JOIN exercises e ON s.exercise_id = e.id 
		WHERE s.profile_id = ? AND e.is_endurance = 1 AND LENGTH(e.content) >= 3000 AND s.wpm >= 70 AND s.accuracy >= 0.97)`, profileID).Scan(&maratonistaAcero)
	if maratonistaAcero {
		unlock("marathon_steel")
	}
	// 23. Control Total
	var controlTotal bool
	db.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM sessions s JOIN exercises e ON s.exercise_id = e.id 
		WHERE s.profile_id = ? AND LENGTH(e.content) >= 500 AND s.accuracy = 1.0)`, profileID).Scan(&controlTotal)
	if controlTotal {
		unlock("total_control")
	}
	// 24. Superación Personal
	// Count exercises where user has beat their own average WPM or previous sessions
	// (Simpler proxy: completed more than 15 sessions and has improved since the first 5)
	if totalSessions >= 15 {
		var avgWpmFirst5, avgWpmLast5 float64
		db.DB.QueryRow("SELECT AVG(wpm) FROM (SELECT wpm FROM sessions WHERE profile_id = ? ORDER BY completed_at ASC LIMIT 5)", profileID).Scan(&avgWpmFirst5)
		db.DB.QueryRow("SELECT AVG(wpm) FROM (SELECT wpm FROM sessions WHERE profile_id = ? ORDER BY completed_at DESC LIMIT 5)", profileID).Scan(&avgWpmLast5)
		if avgWpmLast5 > avgWpmFirst5+10 {
			unlock("personal_best")
		}
	}

	// TIER 4 (Casi Imposibles)
	// 25. El Elegido (The Chosen One)
	if payload.WPM >= 150 && payload.Accuracy == 1.0 {
		unlock("the_chosen_one")
	}
	// 26. Compilador Humano
	var compiladorHumano bool
	db.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM sessions s JOIN exercises e ON s.exercise_id = e.id 
		WHERE s.profile_id = ? AND e.category = 'code' AND LENGTH(e.content) >= 1500 AND s.wpm >= 95 AND s.accuracy = 1.0)`, profileID).Scan(&compiladorHumano)
	if compiladorHumano && payload.BackspacesUsed == 0 {
		unlock("human_compiler")
	}
	// 27. Leyenda de la Tropa
	if streak >= 365 {
		unlock("streak_legend")
	}
	// 28. Deidad de Resistencia
	var deidadResistencia bool
	db.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM sessions s JOIN exercises e ON s.exercise_id = e.id 
		WHERE s.profile_id = ? AND e.is_endurance = 1 AND LENGTH(e.content) >= 5000 AND s.wpm >= 110 AND s.accuracy >= 0.99)`, profileID).Scan(&deidadResistencia)
	if deidadResistencia {
		unlock("endurance_deity")
	}
	// 29. Zen Absoluto
	if consecutivePerfect >= 20 {
		unlock("absolute_zen")
	}
	// 30. El Programador Perfecto
	var progPerfecto bool
	db.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM sessions s JOIN exercises e ON s.exercise_id = e.id 
		WHERE s.profile_id = ? AND e.category = 'code' AND LENGTH(e.content) >= 800 AND s.wpm >= 100 AND s.accuracy = 1.0)`, profileID).Scan(&progPerfecto)
	if progPerfecto {
		unlock("perfect_programmer")
	}

	// --- EXPANSION: 31-45 ACHIEVEMENTS ---
	// 31. Master Classic Spanish (at least 10 Spanish lessons in Classic mode)
	var classicSpaCount int
	db.DB.QueryRow(`SELECT COUNT(DISTINCT s.exercise_id) FROM sessions s JOIN exercises e ON s.exercise_id = e.id 
		WHERE s.profile_id = ? AND e.category = 'spanish' AND s.mode != 'arcade'`, profileID).Scan(&classicSpaCount)
	if classicSpaCount >= 10 {
		unlock("master_classic_spa")
	}

	// 32. Master Classic English (at least 10 English lessons in Classic mode)
	var classicEngCount int
	db.DB.QueryRow(`SELECT COUNT(DISTINCT s.exercise_id) FROM sessions s JOIN exercises e ON s.exercise_id = e.id 
		WHERE s.profile_id = ? AND e.category = 'english' AND s.mode != 'arcade'`, profileID).Scan(&classicEngCount)
	if classicEngCount >= 10 {
		unlock("master_classic_eng")
	}

	// 33. Master Classic Code (at least 10 Code lessons in Classic mode)
	var classicCodeCount int
	db.DB.QueryRow(`SELECT COUNT(DISTINCT s.exercise_id) FROM sessions s JOIN exercises e ON s.exercise_id = e.id 
		WHERE s.profile_id = ? AND e.category = 'code' AND s.mode != 'arcade'`, profileID).Scan(&classicCodeCount)
	if classicCodeCount >= 10 {
		unlock("master_classic_code")
	}

	// 34. Master Classic Numbers (at least 10 Numbers lessons in Classic mode)
	var classicNumCount int
	db.DB.QueryRow(`SELECT COUNT(DISTINCT s.exercise_id) FROM sessions s JOIN exercises e ON s.exercise_id = e.id 
		WHERE s.profile_id = ? AND e.category = 'numbers' AND s.mode != 'arcade'`, profileID).Scan(&classicNumCount)
	if classicNumCount >= 10 {
		unlock("master_classic_num")
	}

	// 35. Master Classic Symbols (at least 10 Symbols lessons in Classic mode)
	var classicSymCount int
	db.DB.QueryRow(`SELECT COUNT(DISTINCT s.exercise_id) FROM sessions s JOIN exercises e ON s.exercise_id = e.id 
		WHERE s.profile_id = ? AND e.category = 'symbols' AND s.mode != 'arcade'`, profileID).Scan(&classicSymCount)
	if classicSymCount >= 10 {
		unlock("master_classic_sym")
	}

	// 36. Master Arcade Spanish (at least 10 Spanish lessons in Arcade mode)
	var arcadeSpaCount int
	db.DB.QueryRow(`SELECT COUNT(DISTINCT s.exercise_id) FROM sessions s JOIN exercises e ON s.exercise_id = e.id 
		WHERE s.profile_id = ? AND e.category = 'spanish' AND s.mode = 'arcade'`, profileID).Scan(&arcadeSpaCount)
	if arcadeSpaCount >= 10 {
		unlock("master_arcade_spa")
	}

	// 37. Master Arcade English (at least 10 English lessons in Arcade mode)
	var arcadeEngCount int
	db.DB.QueryRow(`SELECT COUNT(DISTINCT s.exercise_id) FROM sessions s JOIN exercises e ON s.exercise_id = e.id 
		WHERE s.profile_id = ? AND e.category = 'english' AND s.mode = 'arcade'`, profileID).Scan(&arcadeEngCount)
	if arcadeEngCount >= 10 {
		unlock("master_arcade_eng")
	}

	// 38. Master Arcade Code (at least 10 Code lessons in Arcade mode)
	var arcadeCodeCount int
	db.DB.QueryRow(`SELECT COUNT(DISTINCT s.exercise_id) FROM sessions s JOIN exercises e ON s.exercise_id = e.id 
		WHERE s.profile_id = ? AND e.category = 'code' AND s.mode = 'arcade'`, profileID).Scan(&arcadeCodeCount)
	if arcadeCodeCount >= 10 {
		unlock("master_arcade_code")
	}

	// 39. Master Arcade Numbers (at least 10 Numbers lessons in Arcade mode)
	var arcadeNumCount int
	db.DB.QueryRow(`SELECT COUNT(DISTINCT s.exercise_id) FROM sessions s JOIN exercises e ON s.exercise_id = e.id 
		WHERE s.profile_id = ? AND e.category = 'numbers' AND s.mode = 'arcade'`, profileID).Scan(&arcadeNumCount)
	if arcadeNumCount >= 10 {
		unlock("master_arcade_num")
	}

	// 40. Master Arcade Symbols (at least 10 Symbols lessons in Arcade mode)
	var arcadeSymCount int
	db.DB.QueryRow(`SELECT COUNT(DISTINCT s.exercise_id) FROM sessions s JOIN exercises e ON s.exercise_id = e.id 
		WHERE s.profile_id = ? AND e.category = 'symbols' AND s.mode = 'arcade'`, profileID).Scan(&arcadeSymCount)
	if arcadeSymCount >= 10 {
		unlock("master_arcade_sym")
	}

	// 41. Clean Code Pro (Code >80 WPM with 100% Accuracy)
	var cleanCodePro bool
	db.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM sessions s JOIN exercises e ON s.exercise_id = e.id 
		WHERE s.profile_id = ? AND e.category = 'code' AND s.wpm >= 80 AND s.accuracy = 1.0)`, profileID).Scan(&cleanCodePro)
	if cleanCodePro {
		unlock("elite_code_speed")
	}

	// 42. Elite Spanish Speed (Spanish >110 WPM)
	var eliteSpaSpeed bool
	db.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM sessions s JOIN exercises e ON s.exercise_id = e.id 
		WHERE s.profile_id = ? AND e.category = 'spanish' AND s.wpm >= 110)`, profileID).Scan(&eliteSpaSpeed)
	if eliteSpaSpeed {
		unlock("elite_spanish_speed")
	}

	// 43. Elite English Speed (English >110 WPM)
	var eliteEngSpeed bool
	db.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM sessions s JOIN exercises e ON s.exercise_id = e.id 
		WHERE s.profile_id = ? AND e.category = 'english' AND s.wpm >= 110)`, profileID).Scan(&eliteEngSpeed)
	if eliteEngSpeed {
		unlock("elite_english_speed")
	}

	// 44. Endurance Iron (Complete 3 endurance sessions in a single day with accuracy >= 98%)
	var enduranceIronCount int
	db.DB.QueryRow(`SELECT COUNT(*) FROM sessions s JOIN exercises e ON s.exercise_id = e.id 
		WHERE s.profile_id = ? AND e.is_endurance = 1 AND s.accuracy >= 0.98 AND DATE(s.completed_at) = DATE('now')`, profileID).Scan(&enduranceIronCount)
	if enduranceIronCount >= 3 {
		unlock("elite_endurance")
	}

	// 45. Supreme: Deity of the Tropa (Unlock at least 40 achievements)
	var unlockedCount int
	db.DB.QueryRow("SELECT COUNT(*) FROM achievements WHERE profile_id = ?", profileID).Scan(&unlockedCount)
	if unlockedCount >= 40 {
		unlock("deity_tropa")
	}

	return newlyUnlocked
}

// GetStats handles GET /api/profiles/{id}/stats
func GetStats(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid profile ID")
		return
	}

	var stats models.SummaryStats
	stats.Achievements = []string{}
	stats.ProgressWPM = []models.SessionWPM{}
	stats.WeakestKeys = []models.KeyStats{}

	// Aggregates
	err = db.DB.QueryRow(`SELECT 
		COUNT(id), 
		IFNULL(AVG(wpm), 0), 
		IFNULL(AVG(accuracy), 0), 
		IFNULL(SUM(duration_seconds), 0) 
		FROM sessions WHERE profile_id = ?`, id).
		Scan(&stats.TotalSessions, &stats.AverageWPM, &stats.AverageAccuracy, &stats.TotalDuration)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Weakest keys (highest error rates, min 10 attempts)
	rows, err := db.DB.Query(`SELECT char, CAST(errors AS REAL) / CAST(attempts AS REAL) AS rate 
		FROM key_metrics 
		WHERE profile_id = ? AND attempts >= 10 
		ORDER BY rate DESC LIMIT 5`, id)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var ks models.KeyStats
			if rows.Scan(&ks.Char, &ks.ErrorRate) == nil {
				stats.WeakestKeys = append(stats.WeakestKeys, ks)
			}
		}
	}

	// Progress WPM (last 20 runs)
	pRows, err := db.DB.Query(`SELECT completed_at, wpm 
		FROM sessions 
		WHERE profile_id = ? 
		ORDER BY completed_at ASC LIMIT 20`, id)
	if err == nil {
		defer pRows.Close()
		for pRows.Next() {
			var sw models.SessionWPM
			var completedStr string
			if pRows.Scan(&completedStr, &sw.WPM) == nil {
				sw.CompletedAt, _ = time.Parse("2006-01-02 15:04:05", strings.Split(completedStr, ".")[0])
				stats.ProgressWPM = append(stats.ProgressWPM, sw)
			}
		}
	}

	// Unlocked achievements codes
	aRows, err := db.DB.Query("SELECT code FROM achievements WHERE profile_id = ?", id)
	if err == nil {
		defer aRows.Close()
		for aRows.Next() {
			var code string
			if aRows.Scan(&code) == nil {
				stats.Achievements = append(stats.Achievements, code)
			}
		}
	}

	respondWithJSON(w, http.StatusOK, stats)
}

// GetFailedAttempts handles GET /api/profiles/{id}/retries
func GetFailedAttempts(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid profile ID")
		return
	}

	rows, err := db.DB.Query(`SELECT f.exercise_id, e.title, e.category, e.difficulty, f.wpm, f.accuracy 
		FROM failed_attempts f 
		JOIN exercises e ON f.exercise_id = e.id 
		WHERE f.profile_id = ? 
		ORDER BY f.created_at DESC`, id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	type RetryItem struct {
		ExerciseID string  `json:"exercise_id"`
		Title      string  `json:"title"`
		Category   string  `json:"category"`
		Difficulty string  `json:"difficulty"`
		LastWPM    float64 `json:"last_wpm"`
		LastAcc    float64 `json:"last_accuracy"`
	}

	retries := []RetryItem{}
	for rows.Next() {
		var item RetryItem
		err := rows.Scan(&item.ExerciseID, &item.Title, &item.Category, &item.Difficulty, &item.LastWPM, &item.LastAcc)
		if err == nil {
			retries = append(retries, item)
		}
	}

	respondWithJSON(w, http.StatusOK, retries)
}

// GetAchievements handles GET /api/profiles/{id}/achievements
func GetAchievements(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid profile ID")
		return
	}

	rows, err := db.DB.Query("SELECT code, unlocked_at FROM achievements WHERE profile_id = ?", id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	type AchItem struct {
		Code       string    `json:"code"`
		UnlockedAt time.Time `json:"unlocked_at"`
	}

	unlocked := []AchItem{}
	for rows.Next() {
		var code string
		var unlockedStr string
		if rows.Scan(&code, &unlockedStr) == nil {
			t, _ := time.Parse("2006-01-02 15:04:05", strings.Split(unlockedStr, ".")[0])
			unlocked = append(unlocked, AchItem{Code: code, UnlockedAt: t})
		}
	}

	respondWithJSON(w, http.StatusOK, unlocked)
}
