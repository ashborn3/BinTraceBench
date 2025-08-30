package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteDB struct {
	db   *sql.DB
	path string
}

func NewSQLiteDB(path string) *SQLiteDB {
	return &SQLiteDB{
		path: path,
	}
}

func (s *SQLiteDB) Connect() error {
	db, err := sql.Open("sqlite3", s.path)
	if err != nil {
		return fmt.Errorf("failed to connect to SQLite: %w", err)
	}
	s.db = db
	return nil
}

func (s *SQLiteDB) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *SQLiteDB) Ping() error {
	if s.db == nil {
		return fmt.Errorf("database not connected")
	}
	return s.db.Ping()
}

func (s *SQLiteDB) CreateTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			role TEXT NOT NULL DEFAULT 'user',
			created DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			token TEXT UNIQUE NOT NULL,
			expires DATETIME NOT NULL,
			created DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS analysis_results (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			filename TEXT NOT NULL,
			file_hash TEXT NOT NULL,
			static_data TEXT,
			dynamic_data TEXT,
			created DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS benchmark_results (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			filename TEXT NOT NULL,
			file_hash TEXT NOT NULL,
			result TEXT NOT NULL,
			with_trace BOOLEAN NOT NULL DEFAULT FALSE,
			created DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires)`,
		`CREATE INDEX IF NOT EXISTS idx_analysis_user_hash ON analysis_results(user_id, file_hash)`,
		`CREATE INDEX IF NOT EXISTS idx_benchmark_user_hash ON benchmark_results(user_id, file_hash)`,
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}
	return nil
}

func (s *SQLiteDB) DropTables() error {
	queries := []string{
		"DROP TABLE IF EXISTS benchmark_results",
		"DROP TABLE IF EXISTS analysis_results",
		"DROP TABLE IF EXISTS sessions",
		"DROP TABLE IF EXISTS users",
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			return fmt.Errorf("failed to drop table: %w", err)
		}
	}
	return nil
}

// User management
func (s *SQLiteDB) CreateUser(user *User) error {
	query := `INSERT INTO users (username, password, email, role) VALUES (?, ?, ?, ?)`
	result, err := s.db.Exec(query, user.Username, user.Password, user.Email, user.Role)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get user ID: %w", err)
	}
	user.ID = int(id)
	return nil
}

func (s *SQLiteDB) GetUserByUsername(username string) (*User, error) {
	user := &User{}
	query := `SELECT id, username, password, email, role, created FROM users WHERE username = ?`
	err := s.db.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.Role, &user.Created)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (s *SQLiteDB) GetUserByID(id int) (*User, error) {
	user := &User{}
	query := `SELECT id, username, password, email, role, created FROM users WHERE id = ?`
	err := s.db.QueryRow(query, id).Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.Role, &user.Created)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (s *SQLiteDB) UpdateUser(user *User) error {
	query := `UPDATE users SET username = ?, password = ?, email = ?, role = ? WHERE id = ?`
	_, err := s.db.Exec(query, user.Username, user.Password, user.Email, user.Role, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

func (s *SQLiteDB) DeleteUser(id int) error {
	query := `DELETE FROM users WHERE id = ?`
	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// Session management
func (s *SQLiteDB) CreateSession(session *Session) error {
	query := `INSERT INTO sessions (user_id, token, expires) VALUES (?, ?, ?)`
	result, err := s.db.Exec(query, session.UserID, session.Token, session.Expires)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get session ID: %w", err)
	}
	session.ID = fmt.Sprintf("%d", id)
	return nil
}

func (s *SQLiteDB) GetSessionByToken(token string) (*Session, error) {
	session := &Session{}
	query := `SELECT id, user_id, token, expires, created FROM sessions WHERE token = ? AND expires > ?`
	err := s.db.QueryRow(query, token, time.Now()).Scan(&session.ID, &session.UserID, &session.Token, &session.Expires, &session.Created)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	return session, nil
}

func (s *SQLiteDB) DeleteSession(token string) error {
	query := `DELETE FROM sessions WHERE token = ?`
	_, err := s.db.Exec(query, token)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

func (s *SQLiteDB) DeleteExpiredSessions() error {
	query := `DELETE FROM sessions WHERE expires <= ?`
	_, err := s.db.Exec(query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}
	return nil
}

// Analysis results
func (s *SQLiteDB) SaveAnalysisResult(result *AnalysisResult) error {
	staticData, err := json.Marshal(result.StaticData)
	if err != nil {
		return fmt.Errorf("failed to marshal static data: %w", err)
	}

	dynamicData, err := json.Marshal(result.DynamicData)
	if err != nil {
		return fmt.Errorf("failed to marshal dynamic data: %w", err)
	}

	query := `INSERT INTO analysis_results (user_id, filename, file_hash, static_data, dynamic_data) VALUES (?, ?, ?, ?, ?)`
	dbResult, err := s.db.Exec(query, result.UserID, result.Filename, result.FileHash, string(staticData), string(dynamicData))
	if err != nil {
		return fmt.Errorf("failed to save analysis result: %w", err)
	}

	id, err := dbResult.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get analysis result ID: %w", err)
	}
	result.ID = int(id)
	return nil
}

func (s *SQLiteDB) GetAnalysisResult(id int) (*AnalysisResult, error) {
	result := &AnalysisResult{}
	var staticDataStr, dynamicDataStr string

	query := `SELECT id, user_id, filename, file_hash, static_data, dynamic_data, created FROM analysis_results WHERE id = ?`
	err := s.db.QueryRow(query, id).Scan(&result.ID, &result.UserID, &result.Filename, &result.FileHash, &staticDataStr, &dynamicDataStr, &result.Created)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get analysis result: %w", err)
	}

	if staticDataStr != "" {
		if err := json.Unmarshal([]byte(staticDataStr), &result.StaticData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal static data: %w", err)
		}
	}

	if dynamicDataStr != "" {
		if err := json.Unmarshal([]byte(dynamicDataStr), &result.DynamicData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal dynamic data: %w", err)
		}
	}

	return result, nil
}

func (s *SQLiteDB) GetAnalysisResultsByUser(userID int) ([]*AnalysisResult, error) {
	query := `SELECT id, user_id, filename, file_hash, static_data, dynamic_data, created FROM analysis_results WHERE user_id = ? ORDER BY created DESC`
	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get analysis results: %w", err)
	}
	defer rows.Close()

	var results []*AnalysisResult
	for rows.Next() {
		result := &AnalysisResult{}
		var staticDataStr, dynamicDataStr string

		err := rows.Scan(&result.ID, &result.UserID, &result.Filename, &result.FileHash, &staticDataStr, &dynamicDataStr, &result.Created)
		if err != nil {
			return nil, fmt.Errorf("failed to scan analysis result: %w", err)
		}

		if staticDataStr != "" {
			if err := json.Unmarshal([]byte(staticDataStr), &result.StaticData); err != nil {
				return nil, fmt.Errorf("failed to unmarshal static data: %w", err)
			}
		}

		if dynamicDataStr != "" {
			if err := json.Unmarshal([]byte(dynamicDataStr), &result.DynamicData); err != nil {
				return nil, fmt.Errorf("failed to unmarshal dynamic data: %w", err)
			}
		}

		results = append(results, result)
	}

	return results, nil
}

func (s *SQLiteDB) GetAnalysisResultByHash(userID int, fileHash string) (*AnalysisResult, error) {
	result := &AnalysisResult{}
	var staticDataStr, dynamicDataStr string

	query := `SELECT id, user_id, filename, file_hash, static_data, dynamic_data, created FROM analysis_results WHERE user_id = ? AND file_hash = ? ORDER BY created DESC LIMIT 1`
	err := s.db.QueryRow(query, userID, fileHash).Scan(&result.ID, &result.UserID, &result.Filename, &result.FileHash, &staticDataStr, &dynamicDataStr, &result.Created)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get analysis result: %w", err)
	}

	if staticDataStr != "" {
		if err := json.Unmarshal([]byte(staticDataStr), &result.StaticData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal static data: %w", err)
		}
	}

	if dynamicDataStr != "" {
		if err := json.Unmarshal([]byte(dynamicDataStr), &result.DynamicData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal dynamic data: %w", err)
		}
	}

	return result, nil
}

func (s *SQLiteDB) DeleteAnalysisResult(id int) error {
	query := `DELETE FROM analysis_results WHERE id = ?`
	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete analysis result: %w", err)
	}
	return nil
}

// Benchmark results
func (s *SQLiteDB) SaveBenchmarkResult(result *BenchmarkResult) error {
	resultData, err := json.Marshal(result.Result)
	if err != nil {
		return fmt.Errorf("failed to marshal benchmark result: %w", err)
	}

	query := `INSERT INTO benchmark_results (user_id, filename, file_hash, result, with_trace) VALUES (?, ?, ?, ?, ?)`
	dbResult, err := s.db.Exec(query, result.UserID, result.Filename, result.FileHash, string(resultData), result.WithTrace)
	if err != nil {
		return fmt.Errorf("failed to save benchmark result: %w", err)
	}

	id, err := dbResult.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get benchmark result ID: %w", err)
	}
	result.ID = int(id)
	return nil
}

func (s *SQLiteDB) GetBenchmarkResult(id int) (*BenchmarkResult, error) {
	result := &BenchmarkResult{}
	var resultDataStr string

	query := `SELECT id, user_id, filename, file_hash, result, with_trace, created FROM benchmark_results WHERE id = ?`
	err := s.db.QueryRow(query, id).Scan(&result.ID, &result.UserID, &result.Filename, &result.FileHash, &resultDataStr, &result.WithTrace, &result.Created)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get benchmark result: %w", err)
	}

	if err := json.Unmarshal([]byte(resultDataStr), &result.Result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal benchmark result: %w", err)
	}

	return result, nil
}

func (s *SQLiteDB) GetBenchmarkResultsByUser(userID int) ([]*BenchmarkResult, error) {
	query := `SELECT id, user_id, filename, file_hash, result, with_trace, created FROM benchmark_results WHERE user_id = ? ORDER BY created DESC`
	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get benchmark results: %w", err)
	}
	defer rows.Close()

	var results []*BenchmarkResult
	for rows.Next() {
		result := &BenchmarkResult{}
		var resultDataStr string

		err := rows.Scan(&result.ID, &result.UserID, &result.Filename, &result.FileHash, &resultDataStr, &result.WithTrace, &result.Created)
		if err != nil {
			return nil, fmt.Errorf("failed to scan benchmark result: %w", err)
		}

		if err := json.Unmarshal([]byte(resultDataStr), &result.Result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal benchmark result: %w", err)
		}

		results = append(results, result)
	}

	return results, nil
}

func (s *SQLiteDB) GetBenchmarkResultByHash(userID int, fileHash string) (*BenchmarkResult, error) {
	result := &BenchmarkResult{}
	var resultDataStr string

	query := `SELECT id, user_id, filename, file_hash, result, with_trace, created FROM benchmark_results WHERE user_id = ? AND file_hash = ? ORDER BY created DESC LIMIT 1`
	err := s.db.QueryRow(query, userID, fileHash).Scan(&result.ID, &result.UserID, &result.Filename, &result.FileHash, &resultDataStr, &result.WithTrace, &result.Created)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get benchmark result: %w", err)
	}

	if err := json.Unmarshal([]byte(resultDataStr), &result.Result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal benchmark result: %w", err)
	}

	return result, nil
}

func (s *SQLiteDB) DeleteBenchmarkResult(id int) error {
	query := `DELETE FROM benchmark_results WHERE id = ?`
	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete benchmark result: %w", err)
	}
	return nil
}
