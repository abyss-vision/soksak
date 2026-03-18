package routes

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"soksak/internal/middleware"
	"soksak/internal/services"
	"soksak/internal/storage"
)

const maxUploadBytes = 50 * 1024 * 1024 // 50 MB

// AssetRoutes returns a chi.Router for asset upload and download.
// Must be mounted under /api/companies/{companyUuid}/assets.
func AssetRoutes(assetSvc *services.AssetService, storageSvc storage.StorageProvider) chi.Router {
	r := chi.NewRouter()

	// POST /assets — multipart file upload
	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")

		if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
			middleware.WriteAppError(w, r, middleware.ErrUnprocessable("Failed to parse multipart form"))
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			middleware.WriteAppError(w, r, middleware.ErrUnprocessable("Missing required form field 'file'"))
			return
		}
		defer file.Close()

		body, err := io.ReadAll(io.LimitReader(file, maxUploadBytes+1))
		if err != nil {
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to read uploaded file"))
			return
		}
		if int64(len(body)) > maxUploadBytes {
			middleware.WriteAppError(w, r, middleware.ErrUnprocessable(fmt.Sprintf("File exceeds maximum size of %d bytes", maxUploadBytes)))
			return
		}

		contentType := header.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		sum := sha256.Sum256(body)
		sha256Hex := fmt.Sprintf("%x", sum)

		objectKey := companyUUID + "/assets/" + uuid.NewString()

		if err := storageSvc.Put(r.Context(), objectKey, bytes.NewReader(body)); err != nil {
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to store uploaded file"))
			return
		}

		originalFilename := header.Filename
		var origFilenamePtr *string
		if originalFilename != "" {
			origFilenamePtr = &originalFilename
		}

		asset, err := assetSvc.Create(r.Context(), companyUUID, services.CreateAssetInput{
			Provider:         "local_disk",
			ObjectKey:        objectKey,
			ContentType:      contentType,
			ByteSize:         len(body),
			SHA256:           sha256Hex,
			OriginalFilename: origFilenamePtr,
		})
		if err != nil {
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to record asset"))
			return
		}

		writeJSON(w, http.StatusCreated, asset)
	})

	// GET /assets/{assetUuid} — download asset content
	r.Get("/{assetUuid}", func(w http.ResponseWriter, r *http.Request) {
		companyUUID := chi.URLParam(r, "companyUuid")
		assetUUID := chi.URLParam(r, "assetUuid")

		asset, err := assetSvc.Get(r.Context(), companyUUID, assetUUID)
		if err != nil {
			if appErr, ok := err.(*middleware.AppError); ok {
				middleware.WriteAppError(w, r, appErr)
				return
			}
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to retrieve asset"))
			return
		}

		rc, err := storageSvc.Get(r.Context(), asset.ObjectKey)
		if err != nil {
			middleware.WriteAppError(w, r, middleware.ErrInternal("Failed to read asset from storage"))
			return
		}
		defer rc.Close()

		w.Header().Set("Content-Type", asset.ContentType)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", asset.ByteSize))
		w.Header().Set("Cache-Control", "private, max-age=60")
		w.Header().Set("X-Content-Type-Options", "nosniff")

		filename := "asset"
		if asset.OriginalFilename != nil && *asset.OriginalFilename != "" {
			filename = *asset.OriginalFilename
		}
		w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, filename))

		if _, err := io.Copy(w, rc); err != nil {
			// Headers already sent; log but cannot write error response.
			return
		}
	})

	return r
}
