// internal/handlers/track.go
package handlers

import (
	"github.com/okinrev/veza-web-app/internal/common"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"os/exec"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/okinrev/veza-web-app/internal/database"
	"github.com/okinrev/veza-web-app/internal/middleware"
	"github.com/okinrev/veza-web-app/internal/models"
	"github.com/okinrev/veza-web-app/internal/utils"
)

type TrackHandler struct {
	db *database.DB
}

type TrackResponse struct {
	ID              int      `json:"id"`
	Title           string   `json:"title"`
	Artist          string   `json:"artist"`
	Filename        string   `json:"filename"`
	DurationSeconds *int     `json:"duration_seconds"`
	Tags            []string `json:"tags"`
	IsPublic        bool     `json:"is_public"`
	UploaderID      int      `json:"uploader_id"`
	UploaderName    string   `json:"uploader_name,omitempty"`
	CreatedAt       string   `json:"created_at"`
	UpdatedAt       string   `json:"updated_at"`
	StreamURL       string   `json:"stream_url,omitempty"`
}

type CreateTrackRequest struct {
	Title    string   `json:"title" binding:"required"`
	Artist   string   `json:"artist" binding:"required"`
	Tags     []string `json:"tags"`
	IsPublic bool     `json:"is_public"`
}

type UpdateTrackRequest struct {
	Title    *string   `json:"title,omitempty"`
	Artist   *string   `json:"artist,omitempty"`
	Tags     *[]string `json:"tags,omitempty"`
	IsPublic *bool     `json:"is_public,omitempty"`
}

func NewTrackHandler(db *database.DB) *TrackHandler {
	return &TrackHandler{db: db}
}

// AddTrackWithUpload handles track upload with file
func (h *TrackHandler) AddTrackWithUpload(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	// Parse multipart form
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil { // 32MB max
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid multipart form",
		})
		return
	}

	// Get form values
	title := strings.TrimSpace(c.PostForm("title"))
	artist := strings.TrimSpace(c.PostForm("artist"))
	tagsStr := strings.TrimSpace(c.PostForm("tags"))
	isPublicStr := c.PostForm("is_public")

	if title == "" || artist == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Title and artist are required",
		})
		return
	}

	// Parse tags
	var tags []string
	if tagsStr != "" {
		tags = strings.Split(tagsStr, ",")
		for i := range tags {
			tags[i] = strings.TrimSpace(tags[i])
		}
	}

	// Parse is_public
	isPublic := true // default to public
	if isPublicStr == "false" {
		isPublic = false
	}

	// Get audio file
	file, fileHeader, err := c.Request.FormFile("audio")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Audio file is required",
		})
		return
	}
	defer file.Close()

	// Create audio directory if it doesn't exist
	audioDir := "audio"
	if err := os.MkdirAll(audioDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create audio directory",
		})
		return
	}

	// Save file with safe name
	filename := filepath.Base(fileHeader.Filename)
	savePath := filepath.Join(audioDir, filename)

	out, err := os.Create(savePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to save audio file",
		})
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to write audio file",
		})
		return
	}

	// Insert track into database
	var trackID int
	err = h.db.QueryRow(`
		INSERT INTO tracks (title, artist, filename, duration_seconds, tags, is_public, uploader_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id
	`, title, artist, filename, 0, pq.Array(tags), isPublic, userID).Scan(&trackID)

	if err != nil {
		// Clean up file on database error
		os.Remove(savePath)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to save track to database",
		})
		return
	}

	// Return track data
	track, err := h.getTrackByID(trackID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Track uploaded but failed to retrieve data",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Track uploaded successfully",
		"data":    track,
	})
}

// ListTracks returns a list of tracks
func (h *TrackHandler) ListTracks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	showPrivate := c.Query("show_private") == "true"

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Build query based on permissions
	baseQuery := `
		SELECT t.id, t.title, t.artist, t.filename, t.duration_seconds, t.tags, 
		       t.is_public, t.uploader_id, u.username, t.created_at, t.updated_at
		FROM tracks t
		JOIN users u ON t.uploader_id = u.id
	`
	countQuery := `SELECT COUNT(*) FROM tracks t`

	whereClause := ""
	args := []interface{}{}

	// Apply visibility filters
	if showPrivate {
		// Only show user's own tracks if requesting private
		userID, exists := common.GetUserIDFromContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Authentication required for private tracks",
			})
			return
		}
		whereClause = " WHERE t.uploader_id = $1"
		args = append(args, userID)
	} else {
		// Only public tracks
		whereClause = " WHERE t.is_public = true"
	}

	// Get total count
	var total int
	err := h.db.QueryRow(countQuery+whereClause, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to count tracks",
		})
		return
	}

	// Get tracks
	orderClause := " ORDER BY t.created_at DESC LIMIT $" + strconv.Itoa(len(args)+1) + " OFFSET $" + strconv.Itoa(len(args)+2)
	args = append(args, limit, offset)

	rows, err := h.db.Query(baseQuery+whereClause+orderClause, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve tracks",
		})
		return
	}
	defer rows.Close()

	tracks := []TrackResponse{}
	for rows.Next() {
		var track TrackResponse
		var tags pq.StringArray
		err := rows.Scan(
			&track.ID, &track.Title, &track.Artist, &track.Filename,
			&track.DurationSeconds, &tags, &track.IsPublic, &track.UploaderID,
			&track.UploaderName, &track.CreatedAt, &track.UpdatedAt,
		)
		if err != nil {
			continue
		}
		track.Tags = []string(tags)
		track.StreamURL = fmt.Sprintf("/stream/%s", track.Filename)
		tracks = append(tracks, track)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tracks,
		"meta": gin.H{
			"page":        page,
			"per_page":    limit,
			"total":       total,
			"total_pages": (total + limit - 1) / limit,
		},
	})
}

// GetTrack returns a specific track
func (h *TrackHandler) GetTrack(c *gin.Context) {
	idStr := c.Param("id")
	trackID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid track ID",
		})
		return
	}

	userID, _ := common.GetUserIDFromContext(c)
	track, err := h.getTrackByID(trackID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Track not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    track,
	})
}

// UpdateTrack updates a track's metadata
func (h *TrackHandler) UpdateTrack(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	idStr := c.Param("id")
	trackID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid track ID",
		})
		return
	}

	// Verify ownership
	var ownerID int
	err = h.db.QueryRow("SELECT uploader_id FROM tracks WHERE id = $1", trackID).Scan(&ownerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Track not found",
		})
		return
	}

	if ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Not authorized to update this track",
		})
		return
	}

	var req UpdateTrackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
		})
		return
	}

	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{}
	argCount := 1

	if req.Title != nil {
		setParts = append(setParts, "title = $"+strconv.Itoa(argCount))
		args = append(args, strings.TrimSpace(*req.Title))
		argCount++
	}
	if req.Artist != nil {
		setParts = append(setParts, "artist = $"+strconv.Itoa(argCount))
		args = append(args, strings.TrimSpace(*req.Artist))
		argCount++
	}
	if req.Tags != nil {
		setParts = append(setParts, "tags = $"+strconv.Itoa(argCount))
		args = append(args, pq.Array(*req.Tags))
		argCount++
	}
	if req.IsPublic != nil {
		setParts = append(setParts, "is_public = $"+strconv.Itoa(argCount))
		args = append(args, *req.IsPublic)
		argCount++
	}

	if len(setParts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "No fields to update",
		})
		return
	}

	// Add updated_at and track_id
	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, trackID)

	query := "UPDATE tracks SET " + strings.Join(setParts, ", ") + " WHERE id = $" + strconv.Itoa(argCount)

	_, err = h.db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update track",
		})
		return
	}

	// Return updated track
	track, err := h.getTrackByID(trackID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Track updated but failed to retrieve updated data",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Track updated successfully",
		"data":    track,
	})
}

// DeleteTrack deletes a track and its file
func (h *TrackHandler) DeleteTrack(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	idStr := c.Param("id")
	trackID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid track ID",
		})
		return
	}

	// Get track details for ownership verification and file deletion
	var ownerID int
	var filename string
	err = h.db.QueryRow("SELECT uploader_id, filename FROM tracks WHERE id = $1", trackID).Scan(&ownerID, &filename)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Track not found",
		})
		return
	}

	if ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Not authorized to delete this track",
		})
		return
	}

	// Delete from database first
	_, err = h.db.Exec("DELETE FROM tracks WHERE id = $1", trackID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to delete track from database",
		})
		return
	}

	// Delete file (don't fail if file doesn't exist)
	filePath := filepath.Join("audio", filename)
	os.Remove(filePath)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Track deleted successfully",
	})
}

// StreamAudio serves audio files
func (h *TrackHandler) StreamAudio(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Filename is required",
		})
		return
	}

	// Security: only allow files from audio directory
	safePath := filepath.Join("audio", filepath.Base(filename))
	
	if _, err := os.Stat(safePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Audio file not found",
		})
		return
	}

	c.File(safePath)
}

// GenerateStreamURL generates a signed URL for streaming
func (h *TrackHandler) GenerateStreamURL(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	filename := c.Query("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Filename is required",
		})
		return
	}

	// Generate signed URL (implement your signing logic)
	signedURL, err := utils.GenerateSignedURL(filename, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to generate signed URL",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"url": signedURL,
		},
	})
}

// Helper function to get track by ID with permission checking
func (h *TrackHandler) getTrackByID(trackID, userID int) (*TrackResponse, error) {
	query := `
		SELECT t.id, t.title, t.artist, t.filename, t.duration_seconds, t.tags, 
		       t.is_public, t.uploader_id, u.username, t.created_at, t.updated_at
		FROM tracks t
		JOIN users u ON t.uploader_id = u.id
		WHERE t.id = $1 AND (t.is_public = true OR t.uploader_id = $2)
	`

	var track TrackResponse
	var tags pq.StringArray
	err := h.db.QueryRow(query, trackID, userID).Scan(
		&track.ID, &track.Title, &track.Artist, &track.Filename,
		&track.DurationSeconds, &tags, &track.IsPublic, &track.UploaderID,
		&track.UploaderName, &track.CreatedAt, &track.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	track.Tags = []string(tags)
	track.StreamURL = fmt.Sprintf("/stream/%s", track.Filename)
	return &track, nil
}

// Ajouts pour URLs signées et validation audio

const (
	MaxAudioSize = 100 << 20 // 100MB
	MaxAudioDuration = 600 // 10 minutes in seconds
)

// AudioMetadata represents audio file metadata
type AudioMetadata struct {
	Duration    int    `json:"duration_seconds"`
	Bitrate     int    `json:"bitrate"`
	SampleRate  int    `json:"sample_rate"`
	Format      string `json:"format"`
	Size        int64  `json:"size"`
}

// ValidateAudioFile validates audio file format and metadata
func (h *TrackHandler) validateAudioFile(filePath string) (*AudioMetadata, error) {
	// Use ffprobe to get audio metadata
	cmd := exec.Command("ffprobe", 
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		filePath)
	
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("invalid audio file or ffprobe not available")
	}

	// Parse ffprobe output (simplified - you might want to use a JSON parser)
	metadata := &AudioMetadata{}
	
	// Extract duration using regex (basic implementation)
	durationRegex := regexp.MustCompile(`"duration":"([0-9.]+)"`)
	if matches := durationRegex.FindStringSubmatch(string(output)); len(matches) > 1 {
		if duration, err := strconv.ParseFloat(matches[1], 64); err == nil {
			metadata.Duration = int(duration)
		}
	}

	// Validate duration
	if metadata.Duration > MaxAudioDuration {
		return nil, fmt.Errorf("audio duration exceeds maximum allowed duration of %d seconds", MaxAudioDuration)
	}

	// Extract format
	formatRegex := regexp.MustCompile(`"format_name":"([^"]+)"`)
	if matches := formatRegex.FindStringSubmatch(string(output)); len(matches) > 1 {
		metadata.Format = matches[1]
	}

	// Validate format
	allowedFormats := []string{"mp3", "wav", "flac", "ogg", "m4a", "aac"}
	validFormat := false
	for _, format := range allowedFormats {
		if strings.Contains(metadata.Format, format) {
			validFormat = true
			break
		}
	}
	
	if !validFormat {
		return nil, fmt.Errorf("unsupported audio format: %s", metadata.Format)
	}

	return metadata, nil
}

// GenerateSignedURL creates a signed URL for audio streaming
func (h *TrackHandler) generateSignedURL(filename string, userID int, secretKey string) (string, error) {
	// Set expiration time (1 hour from now)
	expiration := time.Now().Add(time.Hour).Unix()
	
	// Create signature data
	data := fmt.Sprintf("%s:%d:%d", filename, userID, expiration)
	
	// Generate HMAC signature
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(data))
	signature := hex.EncodeToString(mac.Sum(nil))
	
	// Create signed URL
	signedURL := fmt.Sprintf("/stream/signed/%s?expires=%d&sig=%s&uid=%d", 
		filename, expiration, signature, userID)
	
	return signedURL, nil
}

// ValidateSignature validates a signed URL signature
func (h *TrackHandler) validateSignature(filename string, userID int, expiration int64, signature, secretKey string) bool {
	// Check if expired
	if time.Now().Unix() > expiration {
		return false
	}
	
	// Recreate signature
	data := fmt.Sprintf("%s:%d:%d", filename, userID, expiration)
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(data))
	expectedSignature := hex.EncodeToString(mac.Sum(nil))
	
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// AddTrackWithUpload - Version mise à jour avec validation audio
func (h *TrackHandler) AddTrackWithUpload(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	// Parse multipart form with size limit
	if err := c.Request.ParseMultipartForm(MaxAudioSize); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Audio file too large or invalid form",
		})
		return
	}

	// Get form values
	title := strings.TrimSpace(c.PostForm("title"))
	artist := strings.TrimSpace(c.PostForm("artist"))
	tagsStr := strings.TrimSpace(c.PostForm("tags"))
	isPublicStr := c.PostForm("is_public")

	if title == "" || artist == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Title and artist are required",
		})
		return
	}

	// Parse tags
	var tags []string
	if tagsStr != "" {
		tags = strings.Split(tagsStr, ",")
		for i := range tags {
			tags[i] = strings.TrimSpace(tags[i])
		}
	}

	// Parse is_public
	isPublic := true
	if isPublicStr == "false" {
		isPublic = false
	}

	// Get audio file
	file, fileHeader, err := c.Request.FormFile("audio")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Audio file is required",
		})
		return
	}
	defer file.Close()

	// Validate file size
	if fileHeader.Size > MaxAudioSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Audio file too large (max %dMB)", MaxAudioSize/(1<<20)),
		})
		return
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	allowedExts := []string{".mp3", ".wav", ".flac", ".ogg", ".m4a", ".aac"}
	validExt := false
	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			validExt = true
			break
		}
	}
	
	if !validExt {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Unsupported audio format",
		})
		return
	}

	// Create audio directory
	audioDir := "audio"
	if err := os.MkdirAll(audioDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create audio directory",
		})
		return
	}

	// Generate unique filename
	filename := fmt.Sprintf("%d_%d_%s", userID, time.Now().Unix(), 
		strings.ReplaceAll(filepath.Base(fileHeader.Filename), " ", "_"))
	savePath := filepath.Join(audioDir, filename)

	// Save file
	out, err := os.Create(savePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to save audio file",
		})
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		os.Remove(savePath)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to write audio file",
		})
		return
	}

	// Validate and extract audio metadata
	metadata, err := h.validateAudioFile(savePath)
	if err != nil {
		os.Remove(savePath)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Insert track into database with metadata
	var trackID int
	err = h.db.QueryRow(`
		INSERT INTO tracks (title, artist, filename, duration_seconds, tags, is_public, uploader_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id
	`, title, artist, filename, metadata.Duration, pq.Array(tags), isPublic, userID).Scan(&trackID)

	if err != nil {
		os.Remove(savePath)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to save track to database",
		})
		return
	}

	// Return track data
	track, err := h.getTrackByID(trackID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Track uploaded but failed to retrieve data",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Track uploaded successfully",
		"data":    track,
		"metadata": metadata,
	})
}

// StreamAudioSigned serves audio with signed URL validation
func (h *TrackHandler) StreamAudioSigned(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Filename is required",
		})
		return
	}

	// Get signature parameters
	expiresStr := c.Query("expires")
	signature := c.Query("sig")
	userIDStr := c.Query("uid")

	if expiresStr == "" || signature == "" || userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid signed URL parameters",
		})
		return
	}

	// Parse parameters
	expiration, err := strconv.ParseInt(expiresStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid expiration time",
		})
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid user ID",
		})
		return
	}

	// Validate signature (you should get secret key from config)
	secretKey := "your-secret-key" // This should come from environment/config
	if !h.validateSignature(filename, userID, expiration, signature, secretKey) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Invalid or expired signed URL",
		})
		return
	}

	// Verify track exists and user has access
	var trackExists, isPublic bool
	var uploaderID int
	err = h.db.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM tracks WHERE filename = $1), is_public, uploader_id
		FROM tracks WHERE filename = $1
	`, filename).Scan(&trackExists, &isPublic, &uploaderID)

	if !trackExists {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Track not found",
		})
		return
	}

	// Check access permissions
	if !isPublic && uploaderID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Access denied to private track",
		})
		return
	}

	// Serve the audio file
	safePath := filepath.Join("audio", filepath.Base(filename))
	if _, err := os.Stat(safePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Audio file not found",
		})
		return
	}

	// Set appropriate headers for audio streaming
	c.Header("Accept-Ranges", "bytes")
	c.Header("Content-Type", "audio/mpeg") // Adjust based on file type
	c.File(safePath)
}

// GenerateStreamURL generates a signed streaming URL
func (h *TrackHandler) GenerateStreamURL(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	filename := c.Query("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Filename is required",
		})
		return
	}

	// Verify track exists and user has access
	var trackExists, isPublic bool
	var uploaderID int
	err := h.db.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM tracks WHERE filename = $1), is_public, uploader_id
		FROM tracks WHERE filename = $1
	`, filename).Scan(&trackExists, &isPublic, &uploaderID)

	if !trackExists {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Track not found",
		})
		return
	}

	// Check access permissions
	if !isPublic && uploaderID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Access denied to private track",
		})
		return
	}

	// Generate signed URL
	secretKey := "your-secret-key" // This should come from environment/config
	signedURL, err := h.generateSignedURL(filename, userID, secretKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to generate signed URL",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"signed_url": signedURL,
			"expires_in": 3600, // 1 hour
		},
	})
}

// GetTrackStats returns statistics for a track
func (h *TrackHandler) GetTrackStats(c *gin.Context) {
	idStr := c.Param("id")
	trackID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid track ID",
		})
		return
	}

	// Get track statistics (you might want to implement a plays tracking system)
	var stats struct {
		TrackID     int    `json:"track_id"`
		Title       string `json:"title"`
		Artist      string `json:"artist"`
		Duration    int    `json:"duration_seconds"`
		FileSize    int64  `json:"file_size"`
		Format      string `json:"format"`
		CreatedAt   string `json:"created_at"`
	}

	err = h.db.QueryRow(`
		SELECT id, title, artist, COALESCE(duration_seconds, 0), created_at
		FROM tracks WHERE id = $1 AND is_public = true
	`, trackID).Scan(&stats.TrackID, &stats.Title, &stats.Artist, &stats.Duration, &stats.CreatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Track not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}