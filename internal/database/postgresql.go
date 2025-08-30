package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type PostgreSQLDB struct {
	db       *sql.DB
	host     string
	port     int
	user     string
	password string
	dbname   string
	sslmode  string
}

func NewPostgreSQLDB(host string, port int, user, password, dbname, sslmode string) *PostgreSQLDB {
	return &PostgreSQLDB{
		host:     host,
		port:     port,
		user:     user,
		password: password,
		dbname:   dbname,
		sslmode:  sslmode,
	}
}

func (p *PostgreSQLDB) Connect() error {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.host, p.port, p.user, p.password, p.dbname, p.sslmode)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}
	p.db = db
	return nil
}

func (p *PostgreSQLDB) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

func (p *PostgreSQLDB) Ping() error {
	if p.db == nil {
		return fmt.Errorf("database not connected")
	}
	return p.db.Ping()
}

func (p *PostgreSQLDB) CreateTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(255) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			role VARCHAR(50) NOT NULL DEFAULT 'user',
			created TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL,
			token VARCHAR(255) UNIQUE NOT NULL,
			expires TIMESTAMP NOT NULL,
			created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS analysis_results (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL,
			filename VARCHAR(255) NOT NULL,
			file_hash VARCHAR(255) NOT NULL,
			static_data JSONB,
			dynamic_data JSONB,
			created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS benchmark_results (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL,
			filename VARCHAR(255) NOT NULL,
			file_hash VARCHAR(255) NOT NULL,
			result JSONB NOT NULL,
			with_trace BOOLEAN NOT NULL DEFAULT FALSE,
			created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires)`,
		`CREATE INDEX IF NOT EXISTS idx_analysis_user_hash ON analysis_results(user_id, file_hash)`,
		`CREATE INDEX IF NOT EXISTS idx_benchmark_user_hash ON benchmark_results(user_id, file_hash)`,
	}

	for _, query := range queries {
		if _, err := p.db.Exec(query); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}
	return nil
}

func (p *PostgreSQLDB) DropTables() error {
	queries := []string{
		"DROP TABLE IF EXISTS benchmark_results CASCADE",
		"DROP TABLE IF EXISTS analysis_results CASCADE",
		"DROP TABLE IF EXISTS sessions CASCADE",
		"DROP TABLE IF EXISTS users CASCADE",
	}

	for _, query := range queries {
		if _, err := p.db.Exec(query); err != nil {
			return fmt.Errorf("failed to drop table: %w", err)
		}
	}
	return nil
}

// User management
func (p *PostgreSQLDB) CreateUser(user *User) error {
	query := `INSERT INTO users (username, password, email, role) VALUES ($1, $2, $3, $4) RETURNING id, created`
	err := p.db.QueryRow(query, user.Username, user.Password, user.Email, user.Role).Scan(&user.ID, &user.Created)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (p *PostgreSQLDB) GetUserByUsername(username string) (*User, error) {
	user := &User{}
	query := `SELECT id, username, password, email, role, created FROM users WHERE username = $1`
	err := p.db.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.Role, &user.Created)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (p *PostgreSQLDB) GetUserByID(id int) (*User, error) {
	user := &User{}
	query := `SELECT id, username, password, email, role, created FROM users WHERE id = $1`
	err := p.db.QueryRow(query, id).Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.Role, &user.Created)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (p *PostgreSQLDB) UpdateUser(user *User) error {
	query := `UPDATE users SET username = $1, password = $2, email = $3, role = $4 WHERE id = $5`
	_, err := p.db.Exec(query, user.Username, user.Password, user.Email, user.Role, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

func (p *PostgreSQLDB) DeleteUser(id int) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := p.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// Session management
func (p *PostgreSQLDB) CreateSession(session *Session) error {
	query := `INSERT INTO sessions (user_id, token, expires) VALUES ($1, $2, $3) RETURNING id, created`
	var id int
	err := p.db.QueryRow(query, session.UserID, session.Token, session.Expires).Scan(&id, &session.Created)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	session.ID = fmt.Sprintf("%d", id)
	return nil
}

func (p *PostgreSQLDB) GetSessionByToken(token string) (*Session, error) {
	session := &Session{}
	query := `SELECT id, user_id, token, expires, created FROM sessions WHERE token = $1 AND expires > $2`
	err := p.db.QueryRow(query, token, time.Now()).Scan(&session.ID, &session.UserID, &session.Token, &session.Expires, &session.Created)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	return session, nil
}

func (p *PostgreSQLDB) DeleteSession(token string) error {
	query := `DELETE FROM sessions WHERE token = $1`
	_, err := p.db.Exec(query, token)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

func (p *PostgreSQLDB) DeleteExpiredSessions() error {
	query := `DELETE FROM sessions WHERE expires <= $1`
	_, err := p.db.Exec(query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}
	return nil
}

// Analysis results
func (p *PostgreSQLDB) SaveAnalysisResult(result *AnalysisResult) error {
	staticData, err := json.Marshal(result.StaticData)
	if err != nil {
		return fmt.Errorf("failed to marshal static data: %w", err)
	}

	dynamicData, err := json.Marshal(result.DynamicData)
	if err != nil {
		return fmt.Errorf("failed to marshal dynamic data: %w", err)
	}

	query := `INSERT INTO analysis_results (user_id, filename, file_hash, static_data, dynamic_data) VALUES ($1, $2, $3, $4, $5) RETURNING id, created`
	err = p.db.QueryRow(query, result.UserID, result.Filename, result.FileHash, string(staticData), string(dynamicData)).Scan(&result.ID, &result.Created)
	if err != nil {
		return fmt.Errorf("failed to save analysis result: %w", err)
	}
	return nil
}

func (p *PostgreSQLDB) GetAnalysisResult(id int) (*AnalysisResult, error) {
	result := &AnalysisResult{}
	var staticDataStr, dynamicDataStr string

	query := `SELECT id, user_id, filename, file_hash, static_data, dynamic_data, created FROM analysis_results WHERE id = $1`
	err := p.db.QueryRow(query, id).Scan(&result.ID, &result.UserID, &result.Filename, &result.FileHash, &staticDataStr, &dynamicDataStr, &result.Created)
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

func (p *PostgreSQLDB) GetAnalysisResultsByUser(userID int) ([]*AnalysisResult, error) {
	query := `SELECT id, user_id, filename, file_hash, static_data, dynamic_data, created FROM analysis_results WHERE user_id = $1 ORDER BY created DESC`
	rows, err := p.db.Query(query, userID)
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

func (p *PostgreSQLDB) GetAnalysisResultByHash(userID int, fileHash string) (*AnalysisResult, error) {
	result := &AnalysisResult{}
	var staticDataStr, dynamicDataStr string

	query := `SELECT id, user_id, filename, file_hash, static_data, dynamic_data, created FROM analysis_results WHERE user_id = $1 AND file_hash = $2 ORDER BY created DESC LIMIT 1`
	err := p.db.QueryRow(query, userID, fileHash).Scan(&result.ID, &result.UserID, &result.Filename, &result.FileHash, &staticDataStr, &dynamicDataStr, &result.Created)
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

func (p *PostgreSQLDB) DeleteAnalysisResult(id int) error {
	query := `DELETE FROM analysis_results WHERE id = $1`
	_, err := p.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete analysis result: %w", err)
	}
	return nil
}

// Benchmark results
func (p *PostgreSQLDB) SaveBenchmarkResult(result *BenchmarkResult) error {
	resultData, err := json.Marshal(result.Result)
	if err != nil {
		return fmt.Errorf("failed to marshal benchmark result: %w", err)
	}

	query := `INSERT INTO benchmark_results (user_id, filename, file_hash, result, with_trace) VALUES ($1, $2, $3, $4, $5) RETURNING id, created`
	err = p.db.QueryRow(query, result.UserID, result.Filename, result.FileHash, string(resultData), result.WithTrace).Scan(&result.ID, &result.Created)
	if err != nil {
		return fmt.Errorf("failed to save benchmark result: %w", err)
	}
	return nil
}

func (p *PostgreSQLDB) GetBenchmarkResult(id int) (*BenchmarkResult, error) {
	result := &BenchmarkResult{}
	var resultDataStr string

	query := `SELECT id, user_id, filename, file_hash, result, with_trace, created FROM benchmark_results WHERE id = $1`
	err := p.db.QueryRow(query, id).Scan(&result.ID, &result.UserID, &result.Filename, &result.FileHash, &resultDataStr, &result.WithTrace, &result.Created)
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

func (p *PostgreSQLDB) GetBenchmarkResultsByUser(userID int) ([]*BenchmarkResult, error) {
	query := `SELECT id, user_id, filename, file_hash, result, with_trace, created FROM benchmark_results WHERE user_id = $1 ORDER BY created DESC`
	rows, err := p.db.Query(query, userID)
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

func (p *PostgreSQLDB) GetBenchmarkResultByHash(userID int, fileHash string) (*BenchmarkResult, error) {
	result := &BenchmarkResult{}
	var resultDataStr string

	query := `SELECT id, user_id, filename, file_hash, result, with_trace, created FROM benchmark_results WHERE user_id = $1 AND file_hash = $2 ORDER BY created DESC LIMIT 1`
	err := p.db.QueryRow(query, userID, fileHash).Scan(&result.ID, &result.UserID, &result.Filename, &result.FileHash, &resultDataStr, &result.WithTrace, &result.Created)
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

func (p *PostgreSQLDB) DeleteBenchmarkResult(id int) error {
	query := `DELETE FROM benchmark_results WHERE id = $1`
	_, err := p.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete benchmark result: %w", err)
	}
	return nil
}
