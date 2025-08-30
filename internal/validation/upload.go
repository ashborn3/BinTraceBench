package validation

import (
	"fmt"
	"mime/multipart"
	"net/http"
)

const (
	MaxUploadSize = 100 << 20 // 100MB
)

func ValidateFileUpload(r *http.Request) (*multipart.FileHeader, []byte, error) {
	if err := r.ParseMultipartForm(MaxUploadSize); err != nil {
		return nil, nil, fmt.Errorf("failed to parse form: %v", err)
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		return nil, nil, fmt.Errorf("missing file in request")
	}
	defer file.Close()

	if header.Size > MaxUploadSize {
		return nil, nil, fmt.Errorf("file too large: %d bytes (max %d)", header.Size, MaxUploadSize)
	}

	data := make([]byte, header.Size)
	_, err = file.Read(data)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read file: %v", err)
	}

	return header, data, nil
}
