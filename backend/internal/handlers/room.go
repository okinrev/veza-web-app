// internal/handlers/room.go - Handler complet pour les salles de chat

package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"veza-web-app/internal/database"
	"veza-web-app/internal/middleware"
)

type RoomHandler struct {
	db *database.DB
}

type RoomResponse struct {
	ID            int     `json:"id"`
	Name          string  `json:"name"`
	Description   *string `json:"description"`
	IsPrivate     bool    `json:"is_private"`
	CreatorID     *int    `json:"creator_id"`
	CreatorName   string  `json:"creator_name,omitempty"`
	MemberCount   int     `json:"member_count"`
	OnlineCount   int     `json:"online_count,omitempty"`
	LastActivity  *string `json:"last_activity"`
	LastMessage   *string `json:"last_message"`
	UnreadCount   int     `json:"unread_count,omitempty"`
	IsMember      bool    `json:"is_member,omitempty"`
	UserRole      string  `json:"user_role,omitempty"` // "owner", "admin", "member"
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

type CreateRoomRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=50"`
	Description *string `json:"description"`
	IsPrivate   bool    `json:"is_private"`
	Password    *string `json:"password,omitempty"`
}

type UpdateRoomRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	IsPrivate   *bool   `json:"is_private,omitempty"`
	Password    *string `json:"password,omitempty"`
}

type RoomMemberResponse struct {
	UserID      int     `json:"user_id"`
	Username    string  `json:"username"`
	Avatar      *string `json:"avatar"`
	Role        string  `json:"role"` // "owner", "admin", "member"
	JoinedAt    string  `json:"joined_at"`
	LastSeen    *string `json:"last_seen"`
	IsOnline    bool    `json:"is_online"`
	MessageCount int    `json:"message_count"`
}

type JoinRoomRequest struct {
	Password *string `json:"password,omitempty"`
}

type RoomMessageResponse struct {
	ID        int     `json:"id"`
	RoomID    int     `json:"room_id"`
	UserID    int     `json:"user_id"`
	Username  string  `json:"username"`
	Avatar    *string `json:"avatar"`
	Content   string  `json:"content"`
	MessageType string `json:"message_type"` // "message", "join", "leave", "system"
	EditedAt  *string `json:"edited_at"`
	CreatedAt string  `json:"created_at"`
}

func NewRoomHandler(db *database.DB) *RoomHandler {
	return &RoomHandler{db: db}
}

// GetPublicRooms returns list of public rooms
func (h *RoomHandler) GetPublicRooms(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := strings.TrimSpace(c.Query("search"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Build query
	baseQuery := `
		SELECT r.id, r.name, r.description, r.is_private, r.creator_id, 
		       u.username as creator_name, r.created_at, r.updated_at,
		       COALESCE(member_count.count, 0) as member_count,
		       last_msg.content as last_message,
		       last_msg.timestamp as last_activity
		FROM rooms r
		LEFT JOIN users u ON r.creator_id = u.id
		LEFT JOIN (
			SELECT room_id, COUNT(*) as count
			FROM room_members
			GROUP BY room_id
		) member_count ON r.id = member_count.room_id
		LEFT JOIN (
			SELECT DISTINCT ON (room_id) room_id, content, timestamp
			FROM room_messages
			WHERE message_type = 'message'
			ORDER BY room_id, timestamp DESC
		) last_msg ON r.id = last_msg.room_id
		WHERE r.is_private = false
	`

	countQuery := "SELECT COUNT(*) FROM rooms WHERE is_private = false"
	args := []interface{}{}
	argCount := 1

	if search != "" {
		searchClause := " AND (LOWER(r.name) LIKE $" + strconv.Itoa(argCount) + " OR LOWER(r.description) LIKE $" + strconv.Itoa(argCount) + ")"
		baseQuery += searchClause
		countQuery += searchClause
		args = append(args, "%"+strings.ToLower(search)+"%")
		argCount++
	}

	// Get total count
	var total int
	err := h.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to count rooms",
		})
		return
	}

	// Get rooms
	orderClause := " ORDER BY member_count DESC, last_activity DESC NULLS LAST, r.created_at DESC LIMIT $" + strconv.Itoa(argCount) + " OFFSET $" + strconv.Itoa(argCount+1)
	args = append(args, limit, offset)

	rows, err := h.db.Query(baseQuery+orderClause, args...)
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
			&room.CreatorID, &room.CreatorName, &room.CreatedAt, &room.UpdatedAt,
			&room.MemberCount, &room.LastMessage, &room.LastActivity,
		)
		if err != nil {
			continue
		}
		rooms = append(rooms, room)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    rooms,
		"meta": gin.H{
			"page":        page,
			"per_page":    limit,
			"total":       total,
			"total_pages": (total + limit - 1) / limit,
		},
	})
}

// CreateRoom creates a new chat room
func (h *RoomHandler) CreateRoom(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
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

	// Hash password if provided
	var hashedPassword *string
	if req.Password != nil && *req.Password != "" {
		// You should implement password hashing here
		hashedPassword = req.Password // For now, storing plain text (NOT SECURE)
	}

	// Create room
	var roomID int
	err = h.db.QueryRow(`
		INSERT INTO rooms (name, description, is_private, creator_id, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id
	`, req.Name, req.Description, req.IsPrivate, userID, hashedPassword).Scan(&roomID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create room",
		})
		return
	}

	// Add creator as room member with owner role
	_, err = h.db.Exec(`
		INSERT INTO room_members (room_id, user_id, role, joined_at)
		VALUES ($1, $2, 'owner', NOW())
	`, roomID, userID)

	if err != nil {
		// Rollback room creation if member insertion fails
		h.db.Exec("DELETE FROM rooms WHERE id = $1", roomID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to add creator to room",
		})
		return
	}

	// Add system message for room creation
	h.addSystemMessage(roomID, fmt.Sprintf("Room created by %s", getUsernameByID(h.db, userID)))

	// Get created room
	room, err := h.getRoomByID(roomID, userID)
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

// GetRoom returns a specific room
func (h *RoomHandler) GetRoom(c *gin.Context) {
	userID, _ := common.GetUserIDFromContext(c)

	roomIDStr := c.Param("id")
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid room ID",
		})
		return
	}

	room, err := h.getRoomByID(roomID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Room not found or access denied",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    room,
	})
}

// JoinRoom allows user to join a room
func (h *RoomHandler) JoinRoom(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	roomIDStr := c.Param("id")
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid room ID",
		})
		return
	}

	var req JoinRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Password might be optional
		req.Password = nil
	}

	// Check if room exists and get info
	var isPrivate bool
	var passwordHash *string
	err = h.db.QueryRow(`
		SELECT is_private, password_hash FROM rooms WHERE id = $1
	`, roomID).Scan(&isPrivate, &passwordHash)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Room not found",
		})
		return
	}

	// Check password if room is password protected
	if passwordHash != nil && *passwordHash != "" {
		if req.Password == nil || *req.Password != *passwordHash {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid password",
			})
			return
		}
	}

	// Check if user is already a member
	var existingMemberCount int
	err = h.db.QueryRow(`
		SELECT COUNT(*) FROM room_members WHERE room_id = $1 AND user_id = $2
	`, roomID, userID).Scan(&existingMemberCount)

	if err == nil && existingMemberCount > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"error":   "Already a member of this room",
		})
		return
	}

	// Add user to room
	_, err = h.db.Exec(`
		INSERT INTO room_members (room_id, user_id, role, joined_at)
		VALUES ($1, $2, 'member', NOW())
	`, roomID, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to join room",
		})
		return
	}

	// Add system message for user joining
	username := getUsernameByID(h.db, userID)
	h.addSystemMessage(roomID, fmt.Sprintf("%s joined the room", username))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Successfully joined room",
	})
}

// LeaveRoom allows user to leave a room
func (h *RoomHandler) LeaveRoom(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	roomIDStr := c.Param("id")
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid room ID",
		})
		return
	}

	// Check if user is a member and get role
	var userRole string
	err = h.db.QueryRow(`
		SELECT role FROM room_members WHERE room_id = $1 AND user_id = $2
	`, roomID, userID).Scan(&userRole)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Not a member of this room",
		})
		return
	}

	// Room owners cannot leave (they must transfer ownership or delete room)
	if userRole == "owner" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Room owners cannot leave. Transfer ownership or delete the room instead.",
		})
		return
	}

	// Remove user from room
	_, err = h.db.Exec(`
		DELETE FROM room_members WHERE room_id = $1 AND user_id = $2
	`, roomID, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to leave room",
		})
		return
	}

	// Add system message for user leaving
	username := getUsernameByID(h.db, userID)
	h.addSystemMessage(roomID, fmt.Sprintf("%s left the room", username))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Successfully left room",
	})
}

// GetRoomMembers returns members of a room
func (h *RoomHandler) GetRoomMembers(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	roomIDStr := c.Param("id")
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid room ID",
		})
		return
	}

	// Check if user has access to room
	if !h.hasRoomAccess(roomID, userID) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Access denied to this room",
		})
		return
	}

	// Get room members
	rows, err := h.db.Query(`
		SELECT rm.user_id, u.username, u.avatar, rm.role, rm.joined_at,
		       COALESCE(msg_count.count, 0) as message_count
		FROM room_members rm
		JOIN users u ON rm.user_id = u.id
		LEFT JOIN (
			SELECT user_id, COUNT(*) as count
			FROM room_messages
			WHERE room_id = $1 AND message_type = 'message'
			GROUP BY user_id
		) msg_count ON rm.user_id = msg_count.user_id
		WHERE rm.room_id = $1
		ORDER BY rm.role DESC, rm.joined_at ASC
	`, roomID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve room members",
		})
		return
	}
	defer rows.Close()

	members := []RoomMemberResponse{}
	for rows.Next() {
		var member RoomMemberResponse
		err := rows.Scan(
			&member.UserID, &member.Username, &member.Avatar,
			&member.Role, &member.JoinedAt, &member.MessageCount,
		)
		if err != nil {
			continue
		}
		members = append(members, member)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    members,
	})
}

// GetRoomMessages returns messages from a room
func (h *RoomHandler) GetRoomMessages(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	roomIDStr := c.Param("id")
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid room ID",
		})
		return
	}

	// Check if user has access to room
	if !h.hasRoomAccess(roomID, userID) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Access denied to this room",
		})
		return
	}

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
		SELECT rm.id, rm.room_id, rm.user_id, u.username, u.avatar,
		       rm.content, rm.message_type, rm.edited_at, rm.created_at
		FROM room_messages rm
		LEFT JOIN users u ON rm.user_id = u.id
		WHERE rm.room_id = $1
		ORDER BY rm.created_at DESC
		LIMIT $2 OFFSET $3
	`, roomID, limit, offset)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve messages",
		})
		return
	}
	defer rows.Close()

	messages := []RoomMessageResponse{}
	for rows.Next() {
		var message RoomMessageResponse
		err := rows.Scan(
			&message.ID, &message.RoomID, &message.UserID, &message.Username,
			&message.Avatar, &message.Content, &message.MessageType,
			&message.EditedAt, &message.CreatedAt,
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
		},
	})
}

// SendRoomMessage sends a message to a room
func (h *RoomHandler) SendRoomMessage(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	roomIDStr := c.Param("id")
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid room ID",
		})
		return
	}

	// Check if user is a member of the room
	var memberCount int
	err = h.db.QueryRow(`
		SELECT COUNT(*) FROM room_members WHERE room_id = $1 AND user_id = $2
	`, roomID, userID).Scan(&memberCount)

	if err != nil || memberCount == 0 {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Not a member of this room",
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

	// Insert message
	var messageID int
	err = h.db.QueryRow(`
		INSERT INTO room_messages (room_id, user_id, content, message_type, created_at)
		VALUES ($1, $2, $3, 'message', NOW())
		RETURNING id
	`, roomID, userID, strings.TrimSpace(req.Content)).Scan(&messageID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to send message",
		})
		return
	}

	// Update room's last activity
	h.db.Exec("UPDATE rooms SET updated_at = NOW() WHERE id = $1", roomID)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Message sent successfully",
		"data": gin.H{
			"message_id": messageID,
		},
	})
}

// Helper functions
func (h *RoomHandler) getRoomByID(roomID, userID int) (*RoomResponse, error) {
	var room RoomResponse
	err := h.db.QueryRow(`
		SELECT r.id, r.name, r.description, r.is_private, r.creator_id,
		       u.username as creator_name, r.created_at, r.updated_at,
		       COALESCE(member_count.count, 0) as member_count,
		       CASE WHEN rm.user_id IS NOT NULL THEN true ELSE false END as is_member,
		       COALESCE(rm.role, '') as user_role
		FROM rooms r
		LEFT JOIN users u ON r.creator_id = u.id
		LEFT JOIN (
			SELECT room_id, COUNT(*) as count
			FROM room_members
			GROUP BY room_id
		) member_count ON r.id = member_count.room_id
		LEFT JOIN room_members rm ON r.id = rm.room_id AND rm.user_id = $2
		WHERE r.id = $1 AND (r.is_private = false OR rm.user_id IS NOT NULL)
	`, roomID, userID).Scan(
		&room.ID, &room.Name, &room.Description, &room.IsPrivate,
		&room.CreatorID, &room.CreatorName, &room.CreatedAt, &room.UpdatedAt,
		&room.MemberCount, &room.IsMember, &room.UserRole,
	)

	if err != nil {
		return nil, err
	}

	return &room, nil
}

func (h *RoomHandler) hasRoomAccess(roomID, userID int) bool {
	var hasAccess bool
	err := h.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM rooms r
			LEFT JOIN room_members rm ON r.id = rm.room_id AND rm.user_id = $2
			WHERE r.id = $1 AND (r.is_private = false OR rm.user_id IS NOT NULL)
		)
	`, roomID, userID).Scan(&hasAccess)

	return err == nil && hasAccess
}

func (h *RoomHandler) addSystemMessage(roomID int, content string) {
	h.db.Exec(`
		INSERT INTO room_messages (room_id, user_id, content, message_type, created_at)
		VALUES ($1, NULL, $2, 'system', NOW())
	`, roomID, content)
}

func getUsernameByID(db *database.DB, userID int) string {
	var username string
	err := db.QueryRow("SELECT username FROM users WHERE id = $1", userID).Scan(&username)
	if err != nil {
		return "Unknown User"
	}
	return username
}