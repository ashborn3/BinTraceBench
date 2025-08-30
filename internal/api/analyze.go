package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/ashborn3/BinTraceBench/internal/analyzer"
	"github.com/ashborn3/BinTraceBench/internal/auth"
	"github.com/ashborn3/BinTraceBench/internal/database"
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

		r.ParseMultipartForm(100 << 20)

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Missing file in request", http.StatusBadRequest)
			return
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Failed to read file", http.StatusInternalServerError)
			return
		}

		fileHash := auth.GenerateFileHash(data)
		filename := header.Filename

		cached, err := db.GetAnalysisResultByHash(user.ID, fileHash)
		if err == nil && cached != nil {
			// Return cached results
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
