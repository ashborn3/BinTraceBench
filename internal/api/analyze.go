package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/ashborn3/BinTraceBench/internal/analyzer"
)

type AnalyzeResponse struct {
	Static  *analyzer.BinaryInfo           `json:"static"`
	Dynamic []analyzer.VerboseSyscallEntry `json:"dynamic,omitempty"`
}

func AnalyzeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		isDyna := r.URL.Query().Get("dynamic") == "true"

		r.ParseMultipartForm(100 << 20)

		file, _, err := r.FormFile("file")
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

		w.Header().Set("Content-Type", "application/json")
		res := AnalyzeResponse{
			Static:  result,
			Dynamic: dynaResult,
		}
		json.NewEncoder(w).Encode(res)

	}
}
