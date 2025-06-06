// internal/handlers/file.go
package handlers

import (
	"veza-web-app/internal/common"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"veza-web-app/internal/database"
	"veza-web-app/internal/middleware"
	"veza-web-app/internal/models"
)

type FileHandler struct {
	db *database.DB
}

type FileResponse struct {
	ID         int    `json:"id"`
	ProductID  int    `json:"product_id"`
	Filename   string `json:"filename"`
	URL        string `json:"url"`
	Type       string `json:"type"`
	Size       int64  `json:"size,omitempty"`
	MimeType   string `json:"mime_type,omitempty"`
	UploadedAt string `json:"uploaded_at"`
	UpdatedAt  string `json:"updated_at"`
}

type InternalDocResponse struct {
	ID         int     `json:"id"`
	ProductID  int     `json:"product_id"`
	Title      string  `json:"title"`
	Filename   string  `json:"filename"`
	URL        string  `json:"url"`
	Type       string  `json:"type"`
	Size       int64   `json:"size,omitempty"`
	MimeType   string  `json:"mime_type,omitempty"`
	UploadedAt string  `json:"uploaded_at"`
	UpdatedAt  string  `json:"updated_at"`
}

func NewFileHandler(db *database.DB) *FileHandler {
	return &FileHandler{db: db}
}

// UploadFile handles file upload for a user product
func (h *FileHandler) UploadFile(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	productIDStr := c.Param("id")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid product ID",
		})
		return
	}

	// Verify user owns the product
	var ownerID int
	err = h.db.QueryRow("SELECT user_id FROM user_products WHERE id = $1", productID).Scan(&ownerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Product not found",
		})
		return
	}

	if ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Not authorized to upload files for this product",
		})
		return
	}

	// Parse multipart form
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil { // 10MB max
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid multipart form",
		})
		return
	}

	// Get file type from form
	fileType := c.PostForm("type")
	if fileType == "" {
		fileType = "manual" // default type
	}

	// Get file
	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "File is required",
		})
		return
	}
	defer file.Close()

	// Create uploads directory if it doesn't exist
	uploadsDir := "uploads"
	if err := os.MkdirAll(uploadsDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create uploads directory",
		})
		return
	}

	// Generate safe filename with timestamp
	safeName := fmt.Sprintf("%d_%d_%s", productID, time.Now().Unix(), 
		strings.ReplaceAll(filepath.Base(fileHeader.Filename), " ", "_"))
	savePath := filepath.Join(uploadsDir, safeName)

	// Save file
	out, err := os.Create(savePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to save file",
		})
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to write file",
		})
		return
	}

	// Get file info
	fileInfo, _ := out.Stat()
	fileSize := fileInfo.Size()

	// Insert file record into database
	var fileID int
	err = h.db.QueryRow(`
		INSERT INTO files (product_id, filename, url, type, size, mime_type, uploaded_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id
	`, productID, fileHeader.Filename, "/files/"+safeName, fileType, fileSize, fileHeader.Header.Get("Content-Type")).Scan(&fileID)

	if err != nil {
		// Clean up file on database error
		os.Remove(savePath)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to save file record to database",
		})
		return
	}

	// Return file data
	fileResp, err := h.getFileByID(fileID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "File uploaded but failed to retrieve data",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "File uploaded successfully",
		"data":    fileResp,
	})
}

// ListProductFiles returns files for a product
func (h *FileHandler) ListProductFiles(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	productIDStr := c.Param("id")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid product ID",
		})
		return
	}

	// Verify user owns the product
	var ownerID int
	err = h.db.QueryRow("SELECT user_id FROM user_products WHERE id = $1", productID).Scan(&ownerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Product not found",
		})
		return
	}

	if ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Not authorized to view files for this product",
		})
		return
	}

	// Get files
	rows, err := h.db.Query(`
		SELECT id, product_id, filename, url, type, size, mime_type, uploaded_at, updated_at
		FROM files 
		WHERE product_id = $1 
		ORDER BY uploaded_at DESC
	`, productID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve files",
		})
		return
	}
	defer rows.Close()

	files := []FileResponse{}
	for rows.Next() {
		var file FileResponse
		err := rows.Scan(
			&file.ID, &file.ProductID, &file.Filename, &file.URL,
			&file.Type, &file.Size, &file.MimeType, &file.UploadedAt, &file.UpdatedAt,
		)
		if err != nil {
			continue
		}
		files = append(files, file)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    files,
	})
}

// DownloadFile serves a file for download
func (h *FileHandler) DownloadFile(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	fileIDStr := c.Param("id")
	fileID, err := strconv.Atoi(fileIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid file ID",
		})
		return
	}

	// Get file info and verify access
	var file FileResponse
	var ownerID int
	err = h.db.QueryRow(`
		SELECT f.id, f.product_id, f.filename, f.url, f.type, f.size, f.mime_type, 
		       f.uploaded_at, f.updated_at, up.user_id
		FROM files f
		JOIN user_products up ON f.product_id = up.id
		WHERE f.id = $1
	`, fileID).Scan(
		&file.ID, &file.ProductID, &file.Filename, &file.URL,
		&file.Type, &file.Size, &file.MimeType, &file.UploadedAt, &file.UpdatedAt, &ownerID,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "File not found",
		})
		return
	}

	if ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Not authorized to download this file",
		})
		return
	}

	// Get actual file path (security: prevent path traversal)
	filename := filepath.Base(file.URL) // Extract just the filename
	fullPath := filepath.Join("uploads", filename)

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "File not found on disk",
		})
		return
	}

	// Set headers for download
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", file.Filename))
	if file.MimeType != "" {
		c.Header("Content-Type", file.MimeType)
	} else {
		c.Header("Content-Type", "application/octet-stream")
	}

	// Serve file
	c.File(fullPath)
}

// DeleteFile deletes a file
func (h *FileHandler) DeleteFile(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	fileIDStr := c.Param("id")
	fileID, err := strconv.Atoi(fileIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid file ID",
		})
		return
	}

	// Get file info and verify access
	var fileURL string
	var ownerID int
	err = h.db.QueryRow(`
		SELECT f.url, up.user_id
		FROM files f
		JOIN user_products up ON f.product_id = up.id
		WHERE f.id = $1
	`, fileID).Scan(&fileURL, &ownerID)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "File not found",
		})
		return
	}

	if ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Not authorized to delete this file",
		})
		return
	}

	// Delete from database first
	_, err = h.db.Exec("DELETE FROM files WHERE id = $1", fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to delete file record",
		})
		return
	}

	// Delete file from disk (don't fail if file doesn't exist)
	filename := filepath.Base(fileURL)
	filePath := filepath.Join("uploads", filename)
	os.Remove(filePath)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "File deleted successfully",
	})
}

// UploadInternalDoc handles internal document upload
func (h *FileHandler) UploadInternalDoc(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	productIDStr := c.Param("id")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid product ID",
		})
		return
	}

	// Verify user owns the product
	var ownerID int
	err = h.db.QueryRow("SELECT user_id FROM user_products WHERE id = $1", productID).Scan(&ownerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Product not found",
		})
		return
	}

	if ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Not authorized to upload documents for this product",
		})
		return
	}

	// Parse multipart form
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid multipart form",
		})
		return
	}

	// Get form values
	title := strings.TrimSpace(c.PostForm("title"))
	docType := c.PostForm("type")
	if docType == "" {
		docType = "manual"
	}

	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Title is required",
		})
		return
	}

	// Get file
	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "File is required",
		})
		return
	}
	defer file.Close()

	// Create internal_docs directory
	docsDir := "internal_docs"
	if err := os.MkdirAll(docsDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create docs directory",
		})
		return
	}

	// Generate safe filename
	safeName := fmt.Sprintf("%d_%d_%s", productID, time.Now().Unix(), 
		strings.ReplaceAll(filepath.Base(fileHeader.Filename), " ", "_"))
	savePath := filepath.Join(docsDir, safeName)

	// Save file
	out, err := os.Create(savePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to save file",
		})
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to write file",
		})
		return
	}

	// Get file info
	fileInfo, _ := out.Stat()
	fileSize := fileInfo.Size()

	// Insert document record
	var docID int
	err = h.db.QueryRow(`
		INSERT INTO internal_documents (product_id, title, filename, url, type, size, mime_type, uploaded_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id
	`, productID, title, fileHeader.Filename, "/internal_docs/"+safeName, docType, fileSize, fileHeader.Header.Get("Content-Type")).Scan(&docID)

	if err != nil {
		// Clean up file on database error
		os.Remove(savePath)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to save document record to database",
		})
		return
	}

	// Return document data
	doc, err := h.getInternalDocByID(docID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Document uploaded but failed to retrieve data",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Internal document uploaded successfully",
		"data":    doc,
	})
}

// ListInternalDocs returns internal documents for a product
func (h *FileHandler) ListInternalDocs(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	productIDStr := c.Param("id")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid product ID",
		})
		return
	}

	// Verify user owns the product
	var ownerID int
	err = h.db.QueryRow("SELECT user_id FROM user_products WHERE id = $1", productID).Scan(&ownerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Product not found",
		})
		return
	}

	if ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Not authorized to view documents for this product",
		})
		return
	}

	// Get documents
	rows, err := h.db.Query(`
		SELECT id, product_id, title, filename, url, type, size, mime_type, uploaded_at, updated_at
		FROM internal_documents 
		WHERE product_id = $1 
		ORDER BY uploaded_at DESC
	`, productID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve documents",
		})
		return
	}
	defer rows.Close()

	docs := []InternalDocResponse{}
	for rows.Next() {
		var doc InternalDocResponse
		err := rows.Scan(
			&doc.ID, &doc.ProductID, &doc.Title, &doc.Filename, &doc.URL,
			&doc.Type, &doc.Size, &doc.MimeType, &doc.UploadedAt, &doc.UpdatedAt,
		)
		if err != nil {
			continue
		}
		docs = append(docs, doc)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    docs,
	})
}

// ServeInternalDoc serves an internal document for download
func (h *FileHandler) ServeInternalDoc(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	docIDStr := c.Param("id")
	docID, err := strconv.Atoi(docIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid document ID",
		})
		return
	}

	// Get document info and verify access
	var doc InternalDocResponse
	var ownerID int
	err = h.db.QueryRow(`
		SELECT d.id, d.product_id, d.title, d.filename, d.url, d.type, d.size, d.mime_type,
		       d.uploaded_at, d.updated_at, up.user_id
		FROM internal_documents d
		JOIN user_products up ON d.product_id = up.id
		WHERE d.id = $1
	`, docID).Scan(
		&doc.ID, &doc.ProductID, &doc.Title, &doc.Filename, &doc.URL,
		&doc.Type, &doc.Size, &doc.MimeType, &doc.UploadedAt, &doc.UpdatedAt, &ownerID,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Document not found",
		})
		return
	}

	if ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Not authorized to access this document",
		})
		return
	}

	// Get actual file path (security: prevent path traversal)
	filename := filepath.Base(doc.URL)
	fullPath := filepath.Join("internal_docs", filename)

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Document file not found",
		})
		return
	}

	// Set headers for download
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", doc.Filename))
	if doc.MimeType != "" {
		c.Header("Content-Type", doc.MimeType)
	} else {
		c.Header("Content-Type", "application/octet-stream")
	}

	// Serve file
	c.File(fullPath)
}

// Helper functions
func (h *FileHandler) getFileByID(fileID, userID int) (*FileResponse, error) {
	var file FileResponse
	var ownerID int
	err := h.db.QueryRow(`
		SELECT f.id, f.product_id, f.filename, f.url, f.type, f.size, f.mime_type,
		       f.uploaded_at, f.updated_at, up.user_id
		FROM files f
		JOIN user_products up ON f.product_id = up.id
		WHERE f.id = $1 AND up.user_id = $2
	`, fileID, userID).Scan(
		&file.ID, &file.ProductID, &file.Filename, &file.URL,
		&file.Type, &file.Size, &file.MimeType, &file.UploadedAt, &file.UpdatedAt, &ownerID,
	)
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (h *FileHandler) getInternalDocByID(docID, userID int) (*InternalDocResponse, error) {
	var doc InternalDocResponse
	var ownerID int
	err := h.db.QueryRow(`
		SELECT d.id, d.product_id, d.title, d.filename, d.url, d.type, d.size, d.mime_type,
		       d.uploaded_at, d.updated_at, up.user_id
		FROM internal_documents d
		JOIN user_products up ON d.product_id = up.id
		WHERE d.id = $1 AND up.user_id = $2
	`, docID, userID).Scan(
		&doc.ID, &doc.ProductID, &doc.Title, &doc.Filename, &doc.URL,
		&doc.Type, &doc.Size, &doc.MimeType, &doc.UploadedAt, &doc.UpdatedAt, &ownerID,
	)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

// Ajouts pour validation taille

const (
	MaxFileSize = 10 << 20 // 10MB
	MaxInternalDocSize = 50 << 20 // 50MB
)

// ValidateFileSize validates file size based on type
func (h *FileHandler) validateFileSize(size int64, fileType string) error {
	var maxSize int64
	
	switch fileType {
	case "manual", "warranty", "invoice":
		maxSize = MaxInternalDocSize
	default:
		maxSize = MaxFileSize
	}
	
	if size > maxSize {
		return fmt.Errorf("file size exceeds maximum allowed size of %d bytes", maxSize)
	}
	
	return nil
}

// ValidateFileType validates file type based on extension
func (h *FileHandler) validateFileType(filename, expectedType string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	
	allowedExtensions := map[string][]string{
		"manual":   {".pdf", ".doc", ".docx", ".txt"},
		"warranty": {".pdf", ".jpg", ".jpeg", ".png"},
		"invoice":  {".pdf", ".jpg", ".jpeg", ".png"},
		"image":    {".jpg", ".jpeg", ".png", ".gif", ".webp"},
		"document": {".pdf", ".doc", ".docx", ".txt", ".rtf"},
	}
	
	if allowed, exists := allowedExtensions[expectedType]; exists {
		for _, allowedExt := range allowed {
			if ext == allowedExt {
				return nil
			}
		}
		return fmt.Errorf("file type %s not allowed for %s", ext, expectedType)
	}
	
	return nil
}

// UploadFile - Version mise à jour avec validation
func (h *FileHandler) UploadFile(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	productIDStr := c.Param("id")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid product ID",
		})
		return
	}

	// Verify user owns the product
	var ownerID int
	err = h.db.QueryRow("SELECT user_id FROM user_products WHERE id = $1", productID).Scan(&ownerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Product not found",
		})
		return
	}

	if ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Not authorized to upload files for this product",
		})
		return
	}

	// Parse multipart form with size limit
	if err := c.Request.ParseMultipartForm(MaxInternalDocSize); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "File too large or invalid multipart form",
		})
		return
	}

	// Get file type from form
	fileType := c.PostForm("type")
	if fileType == "" {
		fileType = "manual"
	}

	// Get file
	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "File is required",
		})
		return
	}
	defer file.Close()

	// Validate file size
	if err := h.validateFileSize(fileHeader.Size, fileType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Validate file type
	if err := h.validateFileType(fileHeader.Filename, fileType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Create uploads directory if it doesn't exist
	uploadsDir := "uploads"
	if err := os.MkdirAll(uploadsDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create uploads directory",
		})
		return
	}

	// Generate safe filename with timestamp
	safeName := fmt.Sprintf("%d_%d_%s", productID, time.Now().Unix(), 
		strings.ReplaceAll(filepath.Base(fileHeader.Filename), " ", "_"))
	savePath := filepath.Join(uploadsDir, safeName)

	// Save file
	out, err := os.Create(savePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to save file",
		})
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to write file",
		})
		return
	}

	// Get file info
	fileInfo, _ := out.Stat()
	fileSize := fileInfo.Size()

	// Insert file record into database
	var fileID int
	err = h.db.QueryRow(`
		INSERT INTO files (product_id, filename, url, type, size, mime_type, uploaded_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id
	`, productID, fileHeader.Filename, "/files/"+safeName, fileType, fileSize, fileHeader.Header.Get("Content-Type")).Scan(&fileID)

	if err != nil {
		// Clean up file on database error
		os.Remove(savePath)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to save file record to database",
		})
		return
	}

	// Return file data
	fileResp, err := h.getFileByID(fileID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "File uploaded but failed to retrieve data",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "File uploaded successfully",
		"data":    fileResp,
	})
}

// GetFileStats returns file statistics for admin
func (h *FileHandler) GetFileStats(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	var stats struct {
		TotalFiles     int    `json:"total_files"`
		TotalSize      int64  `json:"total_size"`
		FilesByType    map[string]int `json:"files_by_type"`
		AverageSize    int64  `json:"average_size"`
		LargestFile    int64  `json:"largest_file"`
	}

	// Get total files and size for user
	err := h.db.QueryRow(`
		SELECT COUNT(*), COALESCE(SUM(size), 0), COALESCE(AVG(size), 0), COALESCE(MAX(size), 0)
		FROM files f
		JOIN user_products up ON f.product_id = up.id
		WHERE up.user_id = $1
	`, userID).Scan(&stats.TotalFiles, &stats.TotalSize, &stats.AverageSize, &stats.LargestFile)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get file statistics",
		})
		return
	}

	// Get files by type
	rows, err := h.db.Query(`
		SELECT f.type, COUNT(*)
		FROM files f
		JOIN user_products up ON f.product_id = up.id
		WHERE up.user_id = $1
		GROUP BY f.type
	`, userID)

	if err == nil {
		defer rows.Close()
		stats.FilesByType = make(map[string]int)
		for rows.Next() {
			var fileType string
			var count int
			rows.Scan(&fileType, &count)
			stats.FilesByType[fileType] = count
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}