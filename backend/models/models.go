package models

import "time"

// Profile represents a user profile (for La Tropa family members)
type Profile struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	XP           int       `json:"xp"`
	Level        int       `json:"level"`
	Streak       int       `json:"streak"`
	LastActive   string    `json:"last_active"` // YYYY-MM-DD format
	Theme        string    `json:"theme"`
	SoundEnabled bool      `json:"sound_enabled"`
	SwitchType   string    `json:"switch_type"`
	CreatedAt    time.Time `json:"created_at"`
}

// Exercise represents a typing lesson or custom text
type Exercise struct {
	ID                string  `json:"id"`
	Title             string  `json:"title"`
	Content           string  `json:"content"`
	Category          string  `json:"category"`
	Difficulty        string  `json:"difficulty"`
	IsEndurance       bool    `json:"is_endurance"`
	Stage             string  `json:"stage"`
	TargetWPM         float64 `json:"target_wpm"`
	MinAccuracy       float64 `json:"min_accuracy"`
	Unlocked          bool    `json:"unlocked"`
	HighScoreWPM      float64 `json:"high_score_wpm"`
	HighScoreAccuracy float64 `json:"high_score_accuracy"`
	ArcadeWPM         float64 `json:"arcade_wpm"`
	ArcadeAccuracy    float64 `json:"arcade_accuracy"`
	PlayCount         int     `json:"play_count"`
}

// Session represents a completed typing practice run
type Session struct {
	ID              int       `json:"id"`
	ProfileID       int       `json:"profile_id"`
	ExerciseID      string    `json:"exercise_id"`
	Mode            string    `json:"mode"` // 'lesson', 'arcade', 'endurance', 'retry'
	WPM             float64   `json:"wpm"`
	Accuracy        float64   `json:"accuracy"`
	DurationSeconds int       `json:"duration_seconds"`
	RawDataJSON     string    `json:"raw_data_json"` // Time-series telemetry: [{"time":1,"wpm":65,"errors":0}]
	CompletedAt     time.Time `json:"completed_at"`
}

// FailedAttempt logs tests with low performance to queue them for retry
type FailedAttempt struct {
	ProfileID  int       `json:"profile_id"`
	ExerciseID string    `json:"exercise_id"`
	WPM        float64   `json:"wpm"`
	Accuracy   float64   `json:"accuracy"`
	CreatedAt  time.Time `json:"created_at"`
}

// KeyMetric logs latency and error statistics for each keyboard character
type KeyMetric struct {
	ProfileID    int    `json:"profile_id"`
	Char         string `json:"char"`
	Attempts     int    `json:"attempts"`
	Errors       int    `json:"errors"`
	AvgLatencyMS int    `json:"avg_latency_ms"`
}

// Achievement represents a medal unlocked by a profile
type Achievement struct {
	ID         int       `json:"id"`
	ProfileID  int       `json:"profile_id"`
	Code       string    `json:"code"` // Code from our 30 achievements table
	UnlockedAt time.Time `json:"unlocked_at"`
}

// SummaryStats represents aggregated stats for the dashboard
type SummaryStats struct {
	TotalSessions  int             `json:"total_sessions"`
	AverageWPM     float64         `json:"average_wpm"`
	AverageAccuracy float64        `json:"average_accuracy"`
	TotalDuration  int             `json:"total_duration"` // seconds
	WeakestKeys    []KeyStats      `json:"weakest_keys"`
	ProgressWPM    []SessionWPM    `json:"progress_wpm"`
	Achievements   []string        `json:"achievements"`
}

type KeyStats struct {
	Char      string  `json:"char"`
	ErrorRate float64 `json:"error_rate"`
}

type SessionWPM struct {
	CompletedAt time.Time `json:"completed_at"`
	WPM         float64   `json:"wpm"`
}
