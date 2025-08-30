package api

import (
	"encoding/json"
	"net/http"

	"github.com/ashborn3/BinTraceBench/internal/analyzer"
	"github.com/ashborn3/BinTraceBench/internal/auth"
	"github.com/ashborn3/BinTraceBench/internal/database"
	"github.com/ashborn3/BinTraceBench/internal/sandbox"
	"github.com/ashborn3/BinTraceBench/internal/validation"
	"github.com/ashborn3/BinTraceBench/pkg/logging"
)

type AnalyzeResponse struct {
	ID      int                            `json:"id,omitempty"`
	Static  *analyzer.BinaryInfo           `json:"static"`
	Dynamic []analyzer.VerboseSyscallEntry `json:"dynamic,omitempty"`
	Cached  bool                           `json:"cached,omitempty"`
}

func AnalyzeHandler(db database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		isDyna := r.URL.Query().Get("dynamic") == "true"

		header, data, err := validation.ValidateFileUpload(r)
		if err != nil {
			logging.Warn("File upload validation failed", "error", err, "user", user.Username)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Validate binary before analysis
		if err := sandbox.ValidateBinary(data); err != nil {
			http.Error(w, "Binary validation failed: "+err.Error(), http.StatusBadRequest)
			return
		}

		fileHash := auth.GenerateFileHash(data)
		filename := header.Filename

		if cached, err := db.GetAnalysisResultByHash(user.ID, fileHash); err == nil && cached != nil {
			logging.Info("Returning cached analysis", "user", user.Username, "file", filename)
			response := AnalyzeResponse{
				ID:      cached.ID,
				Static:  cached.StaticData,
				Dynamic: cached.DynamicData,
				Cached:  true,
			}

			if isDyna && len(cached.DynamicData) == 0 {
				dynaResult, err := analyzer.TraceBinary(data)
				if err != nil {
					http.Error(w, "Dynamic analysis failed: "+err.Error(), http.StatusInternalServerError)
					return
				}
				response.Dynamic = dynaResult
				response.Cached = false

				cached.DynamicData = dynaResult
				db.SaveAnalysisResult(cached) // This will update if ID exists
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		result, err := analyzer.AnalyzeBinary(data)
		if err != nil {
			http.Error(w, "Failed to analyze binary: "+err.Error(), http.StatusInternalServerError)
			return
		}

		var dynaResult []analyzer.VerboseSyscallEntry
		if isDyna {
			dynaResult, err = analyzer.TraceBinary(data)
			if err != nil {
				http.Error(w, "Dynamic analysis failed: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}

		analysisResult := &database.AnalysisResult{
			UserID:      user.ID,
			Filename:    filename,
			FileHash:    fileHash,
			StaticData:  result,
			DynamicData: dynaResult,
		}

		if err := db.SaveAnalysisResult(analysisResult); err != nil {
			// Log error but don't fail the request
			// The analysis was successful, saving is a bonus
		}

		w.Header().Set("Content-Type", "application/json")
		res := AnalyzeResponse{
			ID:      analysisResult.ID,
			Static:  result,
			Dynamic: dynaResult,
			Cached:  false,
		}
		json.NewEncoder(w).Encode(res)
	}
}
