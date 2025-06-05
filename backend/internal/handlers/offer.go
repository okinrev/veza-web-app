// internal/handlers/offer.go

package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"veza-web-app/internal/database"
	"veza-web-app/internal/middleware"
)

type OfferHandler struct {
	db *database.DB
}

type OfferResponse struct {
	ID                  int     `json:"id"`
	ListingID           int     `json:"listing_id"`
	FromUserID          int     `json:"from_user_id"`
	FromUsername        string  `json:"from_username"`
	FromUserAvatar      *string `json:"from_user_avatar,omitempty"`
	ProposedProductID   int     `json:"proposed_product_id"`
	ProposedProductName string  `json:"proposed_product_name"`
	Message             *string `json:"message"`
	Status              string  `json:"status"`
	CounterOffer        *string `json:"counter_offer,omitempty"`
	ExpiresAt           *string `json:"expires_at,omitempty"`
	CreatedAt           string  `json:"created_at"`
	UpdatedAt           string  `json:"updated_at"`
	
	// Additional info for listing owner
	ListingTitle        string `json:"listing_title,omitempty"`
	ListingDescription  string `json:"listing_description,omitempty"`
	
	// Status tracking
	ViewedAt            *string `json:"viewed_at,omitempty"`
	ResponseDeadline    *string `json:"response_deadline,omitempty"`
}

type CreateOfferRequest struct {
	ProposedProductID int     `json:"proposed_product_id" binding:"required"`
	Message           *string `json:"message"`
	ExpiresIn         int     `json:"expires_in,omitempty"` // Hours until expiration
}

type UpdateOfferRequest struct {
	Status       *string `json:"status,omitempty"` // "accepted", "rejected", "withdrawn"
	CounterOffer *string `json:"counter_offer,omitempty"`
	Message      *string `json:"message,omitempty"`
}

type OfferStats struct {
	TotalOffers     int                    `json:"total_offers"`
	OffersByStatus  map[string]int         `json:"offers_by_status"`
	AcceptanceRate  float64               `json:"acceptance_rate"`
	AverageResponseTime string            `json:"average_response_time"`
	RecentOffers    []OfferResponse        `json:"recent_offers"`
}

func NewOfferHandler(db *database.DB) *OfferHandler {
	return &OfferHandler{db: db}
}

// CreateOffer creates a new offer on a listing
func (h *OfferHandler) CreateOffer(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	listingIDStr := c.Param("listing_id")
	listingID, err := strconv.Atoi(listingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid listing ID",
		})
		return
	}

	var req CreateOfferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
		})
		return
	}

	// Verify listing exists and is open
	var listingOwnerID int
	var listingStatus string
	err = h.db.QueryRow(`
		SELECT user_id, status FROM listings WHERE id = $1
	`, listingID).Scan(&listingOwnerID, &listingStatus)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Listing not found",
		})
		return
	}

	if listingStatus != "open" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Listing is not open for offers",
		})
		return
	}

	if listingOwnerID == userID {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Cannot make offer on your own listing",
		})
		return
	}

	// Verify user owns the proposed product
	var productOwnerID int
	err = h.db.QueryRow(`
		SELECT user_id FROM user_products WHERE id = $1
	`, req.ProposedProductID).Scan(&productOwnerID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Proposed product not found",
		})
		return
	}

	if productOwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Not authorized to offer this product",
		})
		return
	}

	// Check if user already has a pending offer on this listing
	var existingOfferCount int
	err = h.db.QueryRow(`
		SELECT COUNT(*) FROM offers 
		WHERE listing_id = $1 AND from_user_id = $2 AND status = 'pending'
	`, listingID, userID).Scan(&existingOfferCount)

	if err == nil && existingOfferCount > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"error":   "You already have a pending offer on this listing",
		})
		return
	}

	// Calculate expiration time
	var expiresAt *time.Time
	if req.ExpiresIn > 0 && req.ExpiresIn <= 168 { // Max 7 days
		expiry := time.Now().Add(time.Duration(req.ExpiresIn) * time.Hour)
		expiresAt = &expiry
	}

	// Create offer
	var offerID int
	err = h.db.QueryRow(`
		INSERT INTO offers (listing_id, from_user_id, proposed_product_id, message, expires_at, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, 'pending', NOW(), NOW())
		RETURNING id
	`, listingID, userID, req.ProposedProductID, req.Message, expiresAt).Scan(&offerID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create offer",
		})
		return
	}

	// Get created offer
	offer, err := h.getOfferByID(offerID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Offer created but failed to retrieve data",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Offer created successfully",
		"data":    offer,
	})
}

// GetOffer returns a specific offer
func (h *OfferHandler) GetOffer(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	offerIDStr := c.Param("id")
	offerID, err := strconv.Atoi(offerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid offer ID",
		})
		return
	}

	offer, err := h.getOfferByID(offerID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Offer not found or not authorized",
		})
		return
	}

	// Mark as viewed if user is the listing owner
	var listingOwnerID int
	h.db.QueryRow(`
		SELECT l.user_id FROM offers o
		JOIN listings l ON o.listing_id = l.id
		WHERE o.id = $1
	`, offerID).Scan(&listingOwnerID)

	if listingOwnerID == userID {
		h.db.Exec(`
			UPDATE offers SET viewed_at = NOW(), updated_at = NOW() 
			WHERE id = $1 AND viewed_at IS NULL
		`, offerID)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    offer,
	})
}

// UpdateOffer updates an offer (accept, reject, counter-offer)
func (h *OfferHandler) UpdateOffer(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	offerIDStr := c.Param("id")
	offerID, err := strconv.Atoi(offerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid offer ID",
		})
		return
	}

	var req UpdateOfferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
		})
		return
	}

	// Get offer and listing details
	var fromUserID, listingOwnerID, listingID int
	var currentStatus string
	err = h.db.QueryRow(`
		SELECT o.from_user_id, o.status, o.listing_id, l.user_id
		FROM offers o
		JOIN listings l ON o.listing_id = l.id
		WHERE o.id = $1
	`, offerID).Scan(&fromUserID, &currentStatus, &listingID, &listingOwnerID)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Offer not found",
		})
		return
	}

	// Check authorization based on action
	if req.Status != nil {
		switch *req.Status {
		case "accepted", "rejected":
			// Only listing owner can accept/reject
			if userID != listingOwnerID {
				c.JSON(http.StatusForbidden, gin.H{
					"success": false,
					"error":   "Only listing owner can accept or reject offers",
				})
				return
			}
		case "withdrawn":
			// Only offer creator can withdraw
			if userID != fromUserID {
				c.JSON(http.StatusForbidden, gin.H{
					"success": false,
					"error":   "Only offer creator can withdraw offers",
				})
				return
			}
		default:
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid status",
			})
			return
		}
	}

	// Verify offer can be updated
	if currentStatus != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Offer is not pending and cannot be updated",
		})
		return
	}

	// Handle acceptance
	if req.Status != nil && *req.Status == "accepted" {
		// Start transaction
		tx, err := h.db.Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to start transaction",
			})
			return
		}
		defer tx.Rollback()

		// Accept the offer
		_, err = tx.Exec(`
			UPDATE offers SET status = 'accepted', updated_at = NOW() WHERE id = $1
		`, offerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to accept offer",
			})
			return
		}

		// Close the listing
		_, err = tx.Exec(`
			UPDATE listings SET status = 'closed', updated_at = NOW() WHERE id = $1
		`, listingID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to close listing",
			})
			return
		}

		// Reject all other pending offers for this listing
		_, err = tx.Exec(`
			UPDATE offers SET status = 'rejected', updated_at = NOW() 
			WHERE listing_id = $1 AND id != $2 AND status = 'pending'
		`, listingID, offerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to reject other offers",
			})
			return
		}

		// Commit transaction
		if err = tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to commit transaction",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Offer accepted successfully",
		})
		return
	}

	// Handle regular updates (reject, withdraw, counter-offer)
	setParts := []string{}
	args := []interface{}{}
	argCount := 1

	if req.Status != nil {
		setParts = append(setParts, "status = $"+strconv.Itoa(argCount))
		args = append(args, *req.Status)
		argCount++
	}

	if req.CounterOffer != nil {
		setParts = append(setParts, "counter_offer = $"+strconv.Itoa(argCount))
		args = append(args, *req.CounterOffer)
		argCount++
	}

	if req.Message != nil {
		setParts = append(setParts, "message = $"+strconv.Itoa(argCount))
		args = append(args, *req.Message)
		argCount++
	}

	if len(setParts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "No fields to update",
		})
		return
	}

	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, offerID)

	query := "UPDATE offers SET " + strings.Join(setParts, ", ") + " WHERE id = $" + strconv.Itoa(argCount)

	_, err = h.db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update offer",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Offer updated successfully",
	})
}

// GetUserOffers returns offers made by or received by the user
func (h *OfferHandler) GetUserOffers(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	offerType := c.DefaultQuery("type", "all") // "sent", "received", "all"
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Build query based on type
	baseQuery := `
		SELECT o.id, o.listing_id, o.from_user_id, u.username, u.avatar,
		       o.proposed_product_id, p.name, o.message, o.status, o.counter_offer,
		       o.expires_at, o.created_at, o.updated_at, o.viewed_at,
		       l.description as listing_title
		FROM offers o
		JOIN users u ON o.from_user_id = u.id
		JOIN user_products up ON o.proposed_product_id = up.id
		JOIN products p ON up.product_id = p.id
		JOIN listings l ON o.listing_id = l.id
	`

	countQuery := "SELECT COUNT(*) FROM offers o JOIN listings l ON o.listing_id = l.id"
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argCount := 1

	switch offerType {
	case "sent":
		whereClause += " AND o.from_user_id = $" + strconv.Itoa(argCount)
		args = append(args, userID)
		argCount++
	case "received":
		whereClause += " AND l.user_id = $" + strconv.Itoa(argCount)
		args = append(args, userID)
		argCount++
	default: // "all"
		whereClause += " AND (o.from_user_id = $" + strconv.Itoa(argCount) + " OR l.user_id = $" + strconv.Itoa(argCount) + ")"
		args = append(args, userID)
		argCount++
	}

	if status != "" {
		whereClause += " AND o.status = $" + strconv.Itoa(argCount)
		args = append(args, status)
		argCount++
	}

	// Get total count
	var total int
	err := h.db.QueryRow(countQuery+" "+whereClause, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to count offers",
		})
		return
	}

	// Get offers
	orderClause := " ORDER BY o.created_at DESC LIMIT $" + strconv.Itoa(argCount) + " OFFSET $" + strconv.Itoa(argCount+1)
	args = append(args, limit, offset)

	rows, err := h.db.Query(baseQuery+" "+whereClause+orderClause, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve offers",
		})
		return
	}
	defer rows.Close()

	offers := []OfferResponse{}
	for rows.Next() {
		var offer OfferResponse
		err := rows.Scan(
			&offer.ID, &offer.ListingID, &offer.FromUserID, &offer.FromUsername,
			&offer.FromUserAvatar, &offer.ProposedProductID, &offer.ProposedProductName,
			&offer.Message, &offer.Status, &offer.CounterOffer, &offer.ExpiresAt,
			&offer.CreatedAt, &offer.UpdatedAt, &offer.ViewedAt, &offer.ListingTitle,
		)
		if err != nil {
			continue
		}
		offers = append(offers, offer)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    offers,
		"meta": gin.H{
			"page":        page,
			"per_page":    limit,
			"total":       total,
			"total_pages": (total + limit - 1) / limit,
		},
	})
}

// GetOfferStats returns offer statistics for the user
func (h *OfferHandler) GetOfferStats(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	var stats OfferStats

	// Get total offers
	err := h.db.QueryRow(`
		SELECT COUNT(*) FROM offers o
		JOIN listings l ON o.listing_id = l.id
		WHERE o.from_user_id = $1 OR l.user_id = $1
	`, userID).Scan(&stats.TotalOffers)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get offer statistics",
		})
		return
	}

	// Get offers by status
	statusRows, err := h.db.Query(`
		SELECT o.status, COUNT(*)
		FROM offers o
		JOIN listings l ON o.listing_id = l.id
		WHERE o.from_user_id = $1 OR l.user_id = $1
		GROUP BY o.status
	`, userID)

	if err == nil {
		defer statusRows.Close()
		stats.OffersByStatus = make(map[string]int)
		for statusRows.Next() {
			var status string
			var count int
			statusRows.Scan(&status, &count)
			stats.OffersByStatus[status] = count
		}
	}

	// Calculate acceptance rate (for received offers)
	var acceptedCount, totalReceivedCount int
	h.db.QueryRow(`
		SELECT COUNT(CASE WHEN o.status = 'accepted' THEN 1 END),
		       COUNT(*)
		FROM offers o
		JOIN listings l ON o.listing_id = l.id
		WHERE l.user_id = $1
	`, userID).Scan(&acceptedCount, &totalReceivedCount)

	if totalReceivedCount > 0 {
		stats.AcceptanceRate = float64(acceptedCount) / float64(totalReceivedCount) * 100
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// Helper function to get offer by ID with access control
func (h *OfferHandler) getOfferByID(offerID, userID int) (*OfferResponse, error) {
	var offer OfferResponse
	err := h.db.QueryRow(`
		SELECT o.id, o.listing_id, o.from_user_id, u.username, u.avatar,
		       o.proposed_product_id, p.name, o.message, o.status, o.counter_offer,
		       o.expires_at, o.created_at, o.updated_at, o.viewed_at,
		       l.description as listing_title
		FROM offers o
		JOIN users u ON o.from_user_id = u.id
		JOIN user_products up ON o.proposed_product_id = up.id
		JOIN products p ON up.product_id = p.id
		JOIN listings l ON o.listing_id = l.id
		WHERE o.id = $1 AND (o.from_user_id = $2 OR l.user_id = $2)
	`, offerID, userID).Scan(
		&offer.ID, &offer.ListingID, &offer.FromUserID, &offer.FromUsername,
		&offer.FromUserAvatar, &offer.ProposedProductID, &offer.ProposedProductName,
		&offer.Message, &offer.Status, &offer.CounterOffer, &offer.ExpiresAt,
		&offer.CreatedAt, &offer.UpdatedAt, &offer.ViewedAt, &offer.ListingTitle,
	)

	if err != nil {
		return nil, err
	}

	return &offer, nil
}