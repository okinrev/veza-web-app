// internal/handlers/chat.go
package handlers

import (
	"github.com/okinrev/veza-web-app/internal/common"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/okinrev/veza-web-app/internal/database"
	"github.com/okinrev/veza-web-app/internal/middleware"
	"github.com/okinrev/veza-web-app/internal/models"
)

type ChatHandler struct {
	db *database.DB
}

type MessageResponse struct {
	ID        int     `json:"id"`
	FromUser  int     `json:"from_user"`
	ToUser    *int    `json:"to_user,omitempty"`
	Room      *string `json:"room,omitempty"`
	Content   string  `json:"content"`
	Timestamp string  `json:"timestamp"`
	Username  string  `json:"username"`
	IsRead    bool    `json:"is_read"`
}

type RoomResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description *string `json:"description,omitempty"`
	IsPrivate   bool   `json:"is_private"`
	CreatorID   *int   `json:"creator_id,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type CreateRoomRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=50"`
	Description *string `json:"description,omitempty"`
	IsPrivate   bool    `json:"is_private"`
}

type SendMessageRequest struct {
	Content string `json:"content" binding:"required,min=1,max=1000"`
}

type ConversationSummary struct {
	UserID       int    `json:"user_id"`
	Username     string `json:"username"`
	Avatar       *string `json:"avatar,omitempty"`
	LastMessage  string `json:"last_message"`
	LastActivity string `json:"last_activity"`
	UnreadCount  int    `json:"unread_count"`
}

func NewChatHandler(db *database.DB) *ChatHandler {
	return &ChatHandler{db: db}
}

// GetDirectMessages returns direct message history between two users
func (h *ChatHandler) GetDirectMessages(c *gin.Context) {
	currentUserID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	otherUserIDStr := c.Param("user_id")
	otherUserID, err := strconv.Atoi(otherUserIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid user ID",
		})
		return
	}

	// Verify the other user exists
	var count int
	err = h.db.QueryRow("SELECT COUNT(*) FROM users WHERE id = $1", otherUserID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "User not found",
		})
		return
	}

	// Get pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	offset := (page - 1) * limit

	// Get messages between the two users
	rows, err := h.db.Query(`
		SELECT m.id, m.from_user, m.to_user, m.content, m.timestamp, m.is_read, u.username
		FROM messages m
		JOIN users u ON u.id = m.from_user
		WHERE (m.from_user = $1 AND m.to_user = $2)
		   OR (m.from_user = $2 AND m.to_user = $1)
		ORDER BY m.timestamp DESC
		LIMIT $3 OFFSET $4
	`, currentUserID, otherUserID, limit, offset)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve messages",
		})
		return
	}
	defer rows.Close()

	messages := []MessageResponse{}
	for rows.Next() {
		var message MessageResponse
		err := rows.Scan(
			&message.ID, &message.FromUser, &message.ToUser, &message.Content,
			&message.Timestamp, &message.IsRead, &message.Username,
		)
		if err != nil {
			continue
		}
		messages = append(messages, message)
	}

	// Mark messages as read if they were sent to the current user
	_, err = h.db.Exec(`
		UPDATE messages 
		SET is_read = true, updated_at = NOW() 
		WHERE from_user = $1 AND to_user = $2 AND is_read = false
	`, otherUserID, currentUserID)

	// Note: We don't fail the request if marking as read fails

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    messages,
		"meta": gin.H{
			"page":     page,
			"per_page": limit,
		},
	})
}

// SendDirectMessage sends a direct message to another user
func (h *ChatHandler) SendDirectMessage(c *gin.Context) {
	currentUserID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	otherUserIDStr := c.Param("user_id")
	otherUserID, err := strconv.Atoi(otherUserIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid user ID",
		})
		return
	}

	if currentUserID == otherUserID {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Cannot send message to yourself",
		})
		return
	}

	// Verify the other user exists
	var count int
	err = h.db.QueryRow("SELECT COUNT(*) FROM users WHERE id = $1", otherUserID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "User not found",
		})
		return
	}

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
		})
		return
	}

	// Insert message
	var messageID int
	err = h.db.QueryRow(`
		INSERT INTO messages (from_user, to_user, content, timestamp, is_read, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), false, NOW(), NOW())
		RETURNING id
	`, currentUserID, otherUserID, strings.TrimSpace(req.Content)).Scan(&messageID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to send message",
		})
		return
	}

	// Get the created message with user info
	var message MessageResponse
	err = h.db.QueryRow(`
		SELECT m.id, m.from_user, m.to_user, m.content, m.timestamp, m.is_read, u.username
		FROM messages m
		JOIN users u ON u.id = m.from_user
		WHERE m.id = $1
	`, messageID).Scan(
		&message.ID, &message.FromUser, &message.ToUser, &message.Content,
		&message.Timestamp, &message.IsRead, &message.Username,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Message sent but failed to retrieve data",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Message sent successfully",
		"data":    message,
	})
}

// GetConversations returns a list of conversations for the current user
func (h *ChatHandler) GetConversations(c *gin.Context) {
	currentUserID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Get conversations with last message and unread count
	rows, err := h.db.Query(`
		WITH conversation_partners AS (
			SELECT DISTINCT 
				CASE 
					WHEN from_user = $1 THEN to_user 
					ELSE from_user 
				END as partner_id
			FROM messages 
			WHERE from_user = $1 OR to_user = $1
		),
		last_messages AS (
			SELECT DISTINCT ON (
				CASE 
					WHEN from_user = $1 THEN to_user 
					ELSE from_user 
				END
			)
				CASE 
					WHEN from_user = $1 THEN to_user 
					ELSE from_user 
				END as partner_id,
				content as last_message,
				timestamp as last_activity
			FROM messages 
			WHERE from_user = $1 OR to_user = $1
			ORDER BY 
				CASE 
					WHEN from_user = $1 THEN to_user 
					ELSE from_user 
				END,
				timestamp DESC
		),
		unread_counts AS (
			SELECT from_user as partner_id, COUNT(*) as unread_count
			FROM messages 
			WHERE to_user = $1 AND is_read = false
			GROUP BY from_user
		)
		SELECT 
			cp.partner_id, u.username, u.avatar,
			COALESCE(lm.last_message, '') as last_message,
			COALESCE(lm.last_activity, '1970-01-01 00:00:00') as last_activity,
			COALESCE(uc.unread_count, 0) as unread_count
		FROM conversation_partners cp
		JOIN users u ON cp.partner_id = u.id
		LEFT JOIN last_messages lm ON cp.partner_id = lm.partner_id
		LEFT JOIN unread_counts uc ON cp.partner_id = uc.partner_id
		ORDER BY lm.last_activity DESC NULLS LAST
		LIMIT $2
	`, currentUserID, limit)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve conversations",
		})
		return
	}
	defer rows.Close()

	conversations := []ConversationSummary{}
	for rows.Next() {
		var conv ConversationSummary
		err := rows.Scan(
			&conv.UserID, &conv.Username, &conv.Avatar,
			&conv.LastMessage, &conv.LastActivity, &conv.UnreadCount,
		)
		if err != nil {
			continue
		}
		conversations = append(conversations, conv)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    conversations,
	})
}

// GetPublicRooms returns a list of public chat rooms
func (h *ChatHandler) GetPublicRooms(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit < 1 || limit > 100 {
		limit = 50
	}

	rows, err := h.db.Query(`
		SELECT id, name, description, is_private, creator_id, created_at, updated_at
		FROM rooms 
		WHERE is_private = false 
		ORDER BY created_at DESC
		LIMIT $1
	`, limit)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve rooms",
		})
		return
	}
	defer rows.Close()

	rooms := []RoomResponse{}
	for rows.Next() {
		var room RoomResponse
		err := rows.Scan(
			&room.ID, &room.Name, &room.Description, &room.IsPrivate,
			&room.CreatorID, &room.CreatedAt, &room.UpdatedAt,
		)
		if err != nil {
			continue
		}
		rooms = append(rooms, room)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    rooms,
	})
}

// CreateRoom creates a new chat room
func (h *ChatHandler) CreateRoom(c *gin.Context) {
	currentUserID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	var req CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
		})
		return
	}

	// Check if room name already exists
	var count int
	err := h.db.QueryRow("SELECT COUNT(*) FROM rooms WHERE LOWER(name) = LOWER($1)", req.Name).Scan(&count)
	if err == nil && count > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"error":   "Room name already exists",
		})
		return
	}

	// Create room
	var roomID int
	err = h.db.QueryRow(`
		INSERT INTO rooms (name, description, is_private, creator_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id
	`, strings.TrimSpace(req.Name), req.Description, req.IsPrivate, currentUserID).Scan(&roomID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create room",
		})
		return
	}

	// Get the created room
	var room RoomResponse
	err = h.db.QueryRow(`
		SELECT id, name, description, is_private, creator_id, created_at, updated_at
		FROM rooms WHERE id = $1
	`, roomID).Scan(
		&room.ID, &room.Name, &room.Description, &room.IsPrivate,
		&room.CreatorID, &room.CreatedAt, &room.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Room created but failed to retrieve data",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Room created successfully",
		"data":    room,
	})
}

// GetRoomMessages returns messages from a specific room
func (h *ChatHandler) GetRoomMessages(c *gin.Context) {
	roomName := c.Param("room")
	if roomName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Room name is required",
		})
		return
	}

	// Verify room exists and is accessible
	var roomID int
	var isPrivate bool
	err := h.db.QueryRow("SELECT id, is_private FROM rooms WHERE name = $1", roomName).Scan(&roomID, &isPrivate)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Room not found",
		})
		return
	}

	// For private rooms, ensure user has access (this is a simplified check)
	if isPrivate {
		userID, exists := common.GetUserIDFromContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Authentication required for private rooms",
			})
			return
		}
		// TODO: Implement proper room membership checking
		_ = userID
	}

	// Get pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	offset := (page - 1) * limit

	// Get room messages
	rows, err := h.db.Query(`
		SELECT m.id, m.from_user, m.content, m.timestamp, u.username
		FROM messages m
		JOIN users u ON m.from_user = u.id
		WHERE m.room = $1
		ORDER BY m.timestamp DESC
		LIMIT $2 OFFSET $3
	`, roomName, limit, offset)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve room messages",
		})
		return
	}
	defer rows.Close()

	messages := []MessageResponse{}
	for rows.Next() {
		var message MessageResponse
		message.Room = &roomName
		err := rows.Scan(
			&message.ID, &message.FromUser, &message.Content,
			&message.Timestamp, &message.Username,
		)
		if err != nil {
			continue
		}
		messages = append(messages, message)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    messages,
		"meta": gin.H{
			"page":     page,
			"per_page": limit,
			"room":     roomName,
		},
	})
}

// SendRoomMessage sends a message to a room
func (h *ChatHandler) SendRoomMessage(c *gin.Context) {
	currentUserID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	roomName := c.Param("room")
	if roomName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Room name is required",
		})
		return
	}

	// Verify room exists
	var roomID int
	var isPrivate bool
	err := h.db.QueryRow("SELECT id, is_private FROM rooms WHERE name = $1", roomName).Scan(&roomID, &isPrivate)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Room not found",
		})
		return
	}

	// For private rooms, ensure user has access
	if isPrivate {
		// TODO: Implement proper room membership checking
	}

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
		})
		return
	}

	// Insert room message
	var messageID int
	err = h.db.QueryRow(`
		INSERT INTO messages (from_user, room, content, timestamp, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW(), NOW())
		RETURNING id
	`, currentUserID, roomName, strings.TrimSpace(req.Content)).Scan(&messageID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to send message",
		})
		return
	}

	// Get the created message with user info
	var message MessageResponse
	message.Room = &roomName
	err = h.db.QueryRow(`
		SELECT m.id, m.from_user, m.content, m.timestamp, u.username
		FROM messages m
		JOIN users u ON u.id = m.from_user
		WHERE m.id = $1
	`, messageID).Scan(
		&message.ID, &message.FromUser, &message.Content,
		&message.Timestamp, &message.Username,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Message sent but failed to retrieve data",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Message sent successfully",
		"data":    message,
	})
}

// Ajouts pour système de conversations

// Enhanced ConversationSummary with avatar support
type EnhancedConversationSummary struct {
	UserID       int     `json:"user_id"`
	Username     string  `json:"username"`
	FirstName    *string `json:"first_name,omitempty"`
	LastName     *string `json:"last_name,omitempty"`
	Avatar       *string `json:"avatar,omitempty"`
	LastMessage  string  `json:"last_message"`
	LastActivity string  `json:"last_activity"`
	UnreadCount  int     `json:"unread_count"`
	IsOnline     bool    `json:"is_online,omitempty"`
	LastSeen     *string `json:"last_seen,omitempty"`
}

// MessageWithAvatar includes sender avatar information
type MessageWithAvatar struct {
	ID        int     `json:"id"`
	FromUser  int     `json:"from_user"`
	ToUser    *int    `json:"to_user,omitempty"`
	Room      *string `json:"room,omitempty"`
	Content   string  `json:"content"`
	Timestamp string  `json:"timestamp"`
	Username  string  `json:"username"`
	Avatar    *string `json:"avatar,omitempty"`
	IsRead    bool    `json:"is_read"`
	EditedAt  *string `json:"edited_at,omitempty"`
}

// GetDirectMessages - Version améliorée avec avatars
func (h *ChatHandler) GetDirectMessages(c *gin.Context) {
	currentUserID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	otherUserIDStr := c.Param("user_id")
	otherUserID, err := strconv.Atoi(otherUserIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid user ID",
		})
		return
	}

	// Verify the other user exists
	var count int
	err = h.db.QueryRow("SELECT COUNT(*) FROM users WHERE id = $1 AND role != 'deleted'", otherUserID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "User not found",
		})
		return
	}

	// Get pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	offset := (page - 1) * limit

	// Get messages between the two users with avatar information
	rows, err := h.db.Query(`
		SELECT m.id, m.from_user, m.to_user, m.content, m.timestamp, m.is_read, 
		       u.username, u.avatar, m.edited_at
		FROM messages m
		JOIN users u ON u.id = m.from_user
		WHERE (m.from_user = $1 AND m.to_user = $2)
		   OR (m.from_user = $2 AND m.to_user = $1)
		ORDER BY m.timestamp DESC
		LIMIT $3 OFFSET $4
	`, currentUserID, otherUserID, limit, offset)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve messages",
		})
		return
	}
	defer rows.Close()

	messages := []MessageWithAvatar{}
	for rows.Next() {
		var message MessageWithAvatar
		err := rows.Scan(
			&message.ID, &message.FromUser, &message.ToUser, &message.Content,
			&message.Timestamp, &message.IsRead, &message.Username, &message.Avatar,
			&message.EditedAt,
		)
		if err != nil {
			continue
		}
		messages = append(messages, message)
	}

	// Mark messages as read if they were sent to the current user
	_, err = h.db.Exec(`
		UPDATE messages 
		SET is_read = true, updated_at = NOW() 
		WHERE from_user = $1 AND to_user = $2 AND is_read = false
	`, otherUserID, currentUserID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    messages,
		"meta": gin.H{
			"page":     page,
			"per_page": limit,
		},
	})
}

// GetConversations - Version améliorée avec avatars et détails
func (h *ChatHandler) GetConversations(c *gin.Context) {
	currentUserID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Get conversations with enhanced user information
	rows, err := h.db.Query(`
		WITH conversation_partners AS (
			SELECT DISTINCT 
				CASE 
					WHEN from_user = $1 THEN to_user 
					ELSE from_user 
				END as partner_id
			FROM messages 
			WHERE from_user = $1 OR to_user = $1
		),
		last_messages AS (
			SELECT DISTINCT ON (
				CASE 
					WHEN from_user = $1 THEN to_user 
					ELSE from_user 
				END
			)
				CASE 
					WHEN from_user = $1 THEN to_user 
					ELSE from_user 
				END as partner_id,
				content as last_message,
				timestamp as last_activity
			FROM messages 
			WHERE from_user = $1 OR to_user = $1
			ORDER BY 
				CASE 
					WHEN from_user = $1 THEN to_user 
					ELSE from_user 
				END,
				timestamp DESC
		),
		unread_counts AS (
			SELECT from_user as partner_id, COUNT(*) as unread_count
			FROM messages 
			WHERE to_user = $1 AND is_read = false
			GROUP BY from_user
		)
		SELECT 
			cp.partner_id, u.username, u.first_name, u.last_name, u.avatar,
			COALESCE(lm.last_message, '') as last_message,
			COALESCE(lm.last_activity, '1970-01-01 00:00:00') as last_activity,
			COALESCE(uc.unread_count, 0) as unread_count
		FROM conversation_partners cp
		JOIN users u ON cp.partner_id = u.id
		LEFT JOIN last_messages lm ON cp.partner_id = lm.partner_id
		LEFT JOIN unread_counts uc ON cp.partner_id = uc.partner_id
		WHERE u.role != 'deleted'
		ORDER BY lm.last_activity DESC NULLS LAST
		LIMIT $2
	`, currentUserID, limit)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve conversations",
		})
		return
	}
	defer rows.Close()

	conversations := []EnhancedConversationSummary{}
	for rows.Next() {
		var conv EnhancedConversationSummary
		err := rows.Scan(
			&conv.UserID, &conv.Username, &conv.FirstName, &conv.LastName, &conv.Avatar,
			&conv.LastMessage, &conv.LastActivity, &conv.UnreadCount,
		)
		if err != nil {
			continue
		}
		conversations = append(conversations, conv)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    conversations,
	})
}

// MarkAsRead marks messages as read
func (h *ChatHandler) MarkAsRead(c *gin.Context) {
	currentUserID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	otherUserIDStr := c.Param("user_id")
	otherUserID, err := strconv.Atoi(otherUserIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid user ID",
		})
		return
	}

	// Mark all unread messages from the other user as read
	result, err := h.db.Exec(`
		UPDATE messages 
		SET is_read = true, updated_at = NOW() 
		WHERE from_user = $1 AND to_user = $2 AND is_read = false
	`, otherUserID, currentUserID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to mark messages as read",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Messages marked as read",
		"data": gin.H{
			"messages_marked": rowsAffected,
		},
	})
}

// EditMessage allows users to edit their own messages
func (h *ChatHandler) EditMessage(c *gin.Context) {
	currentUserID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	messageIDStr := c.Param("message_id")
	messageID, err := strconv.Atoi(messageIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid message ID",
		})
		return
	}

	var req struct {
		Content string `json:"content" binding:"required,min=1,max=1000"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
		})
		return
	}

	// Verify user owns the message
	var ownerID int
	var originalContent string
	err = h.db.QueryRow("SELECT from_user, content FROM messages WHERE id = $1", messageID).Scan(&ownerID, &originalContent)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Message not found",
		})
		return
	}

	if ownerID != currentUserID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Not authorized to edit this message",
		})
		return
	}

	// Update message content
	_, err = h.db.Exec(`
		UPDATE messages 
		SET content = $1, edited_at = NOW(), updated_at = NOW() 
		WHERE id = $2
	`, strings.TrimSpace(req.Content), messageID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to edit message",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Message edited successfully",
	})
}

// DeleteMessage allows users to delete their own messages
func (h *ChatHandler) DeleteMessage(c *gin.Context) {
	currentUserID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	messageIDStr := c.Param("message_id")
	messageID, err := strconv.Atoi(messageIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid message ID",
		})
		return
	}

	// Verify user owns the message
	var ownerID int
	err = h.db.QueryRow("SELECT from_user FROM messages WHERE id = $1", messageID).Scan(&ownerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Message not found",
		})
		return
	}

	if ownerID != currentUserID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Not authorized to delete this message",
		})
		return
	}

	// Soft delete: mark as deleted instead of removing from database
	_, err = h.db.Exec(`
		UPDATE messages 
		SET content = '[Message deleted]', edited_at = NOW(), updated_at = NOW() 
		WHERE id = $1
	`, messageID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to delete message",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Message deleted successfully",
	})
}

// GetUnreadCount returns total unread messages count for current user
func (h *ChatHandler) GetUnreadCount(c *gin.Context) {
	currentUserID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	var unreadCount int
	err := h.db.QueryRow(`
		SELECT COUNT(*) 
		FROM messages 
		WHERE to_user = $1 AND is_read = false
	`, currentUserID).Scan(&unreadCount)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get unread count",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"unread_count": unreadCount,
		},
	})
}