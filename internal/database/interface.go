package database

import (
	"time"

	"github.com/ashborn3/BinTraceBench/internal/analyzer"
	"github.com/ashborn3/BinTraceBench/internal/sandbox"
)

type User struct {
	ID       int       `json:"id" db:"id"`
	Username string    `json:"username" db:"username"`
	Password string    `json:"-" db:"password"` // never expose password in JSON
	Email    string    `json:"email" db:"email"`
	Role     string    `json:"role" db:"role"`
	Created  time.Time `json:"created" db:"created"`
}

type Session struct {
	ID      string    `json:"id" db:"id"`
	UserID  int       `json:"user_id" db:"user_id"`
	Token   string    `json:"token" db:"token"`
	Expires time.Time `json:"expires" db:"expires"`
	Created time.Time `json:"created" db:"created"`
}

type AnalysisResult struct {
	ID          int                            `json:"id" db:"id"`
	UserID      int                            `json:"user_id" db:"user_id"`
	Filename    string                         `json:"filename" db:"filename"`
	FileHash    string                         `json:"file_hash" db:"file_hash"`
	StaticData  *analyzer.BinaryInfo           `json:"static_data" db:"static_data"`
	DynamicData []analyzer.VerboseSyscallEntry `json:"dynamic_data" db:"dynamic_data"`
	Created     time.Time                      `json:"created" db:"created"`
}

type BenchmarkResult struct {
	ID        int                  `json:"id" db:"id"`
	UserID    int                  `json:"user_id" db:"user_id"`
	Filename  string               `json:"filename" db:"filename"`
	FileHash  string               `json:"file_hash" db:"file_hash"`
	Result    *sandbox.BenchResult `json:"result" db:"result"`
	WithTrace bool                 `json:"with_trace" db:"with_trace"`
	Created   time.Time            `json:"created" db:"created"`
}

type Database interface {
	// Connection management
	Connect() error
	Close() error
	Ping() error

	// Schema management
	CreateTables() error
	DropTables() error

	// User management
	CreateUser(user *User) error
	GetUserByUsername(username string) (*User, error)
	GetUserByID(id int) (*User, error)
	UpdateUser(user *User) error
	DeleteUser(id int) error

	// Session management
	CreateSession(session *Session) error
	GetSessionByToken(token string) (*Session, error)
	DeleteSession(token string) error
	DeleteExpiredSessions() error

	// Analysis results
	SaveAnalysisResult(result *AnalysisResult) error
	GetAnalysisResult(id int) (*AnalysisResult, error)
	GetAnalysisResultsByUser(userID int) ([]*AnalysisResult, error)
	GetAnalysisResultByHash(userID int, fileHash string) (*AnalysisResult, error)
	DeleteAnalysisResult(id int) error

	// Benchmark results
	SaveBenchmarkResult(result *BenchmarkResult) error
	GetBenchmarkResult(id int) (*BenchmarkResult, error)
	GetBenchmarkResultsByUser(userID int) ([]*BenchmarkResult, error)
	GetBenchmarkResultByHash(userID int, fileHash string) (*BenchmarkResult, error)
	DeleteBenchmarkResult(id int) error
}
