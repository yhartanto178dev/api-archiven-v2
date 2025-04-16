package domain

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/go-playground/validator/v10"
)

var (
	validate = validator.New()

	AllowedMimeTypes = []string{
		"application/pdf",
	}

	MaxFileSize = 10 * 1024 * 1024 // Example: 10 MB
)

func init() {
	validate.RegisterValidation("filename", validateFileName)
	validate.RegisterValidation("mimetype", validateMimeType)
}

type FileUploadRequest struct {
	File interface{} `validate:"required"`
}

func validateFileName(fl validator.FieldLevel) bool {
	filename := fl.Field().String()
	if filename == "" {
		return false
	}

	// Prevent path traversal
	if strings.ContainsAny(filename, "\\/:*?\"<>|") {
		return false
	}

	// Limit extension length
	if len(filepath.Ext(filename)) > 10 {
		return false
	}

	return true
}

func validateMimeType(fl validator.FieldLevel) bool {
	contentType := fl.Field().String()
	if contentType == "" {
		return false
	}

	for _, allowed := range AllowedMimeTypes {
		if contentType == allowed {
			return true
		}
	}
	return false
}

func ValidateUploadRequest(req FileUploadRequest) error {
	return validate.Struct(req)
}

func ValidateMimeType(contentType string) error {
	return validate.Var(contentType, "mimetype")
}

func ValidateFileSize(size int64) error {
	if size > int64(MaxFileSize) {
		return fmt.Errorf("file size exceeds maximum allowed: %s", humanize.Bytes(uint64(MaxFileSize)))
	}
	return nil
}

func FormatValidationErrors(err error) []string {
	var errors []string
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			switch e.Tag() {
			case "required":
				errors = append(errors, fmt.Sprintf("%s is required", e.Field()))
			case "filename":
				errors = append(errors, "invalid file name format")
			case "mimetype":
				errors = append(errors,
					fmt.Sprintf("allowed types: %s",
						strings.Join(AllowedMimeTypes, ", ")))
			case "max":
				errors = append(errors, e.Param())
			}
		}
	}
	return errors
}
