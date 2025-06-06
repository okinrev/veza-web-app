// internal/handlers/offer.go
package handlers

import (
	"github.com/okinrev/veza-web-app/internal/common"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/okinrev/veza-web-app/internal/database"
	"github.com/okinrev/veza-web-app/internal/middleware"
	// "github.com/lib/pq" // Not directly needed here unless manipulating arrays in new ways
)

type OfferHandler struct {
	db *database.DB
}

// OfferResponse struct from the second provided offer.go
type OfferResponse struct {
	ID                  int     `json:"id"`
	ListingID           int     `json:"listing_id"`
	FromUserID          int     `json:"from_user_id"`
	FromUsername        string  `json:"from_username"`
	FromUserAvatar      *string `json:"from_user_avatar,omitempty"`
	ProposedProductID   int     `json:"proposed_product_id"` // This is user_products.id
	ProposedProductName string  `json:"proposed_product_name"`
	Message             *string `json:"message"`
	Status              string  `json:"status"`
	CounterOffer        *string `json:"counter_offer,omitempty"`
	ExpiresAt           *string `json:"expires_at,omitempty"`
	CreatedAt           string  `json:"created_at"`
	UpdatedAt           string  `json:"updated_at"`

	// Additional info
	ListingTitle        *string `json:"listing_title,omitempty"` // Changed to pointer to handle potential NULL from DB
	ListingDescription  *string `json:"listing_description,omitempty"` // Changed to pointer

	// Status tracking
	ViewedAt         *string `json:"viewed_at,omitempty"`
	ResponseDeadline *string `json:"response_deadline,omitempty"` // This field wasn't populated in original helpers, remains so
}

// CreateOfferRequest struct from the second provided offer.go
type CreateOfferRequest struct {
	ProposedProductID int     `json:"proposed_product_id" binding:"required"` // This is user_products.id
	Message           *string `json:"message"`
	ExpiresIn         int     `json:"expires_in,omitempty"` // Hours until expiration
}

// UpdateOfferRequest struct from the second provided offer.go
type UpdateOfferRequest struct {
	Status       *string `json:"status,omitempty"` // "accepted", "rejected", "withdrawn"
	CounterOffer *string `json:"counter_offer,omitempty"`
	Message      *string `json:"message,omitempty"`
}

// OfferStats struct from the second provided offer.go
type OfferStats struct {
	TotalOffers         int            `json:"total_offers"`
	OffersByStatus      map[string]int `json:"offers_by_status"`
	AcceptanceRate      float64        `json:"acceptance_rate"`
	AverageResponseTime string         `json:"average_response_time"` // This field wasn't populated in original GetOfferStats
	RecentOffers        []OfferResponse `json:"recent_offers"` // This field wasn't populated in original GetOfferStats
}

func NewOfferHandler(db *database.DB) *OfferHandler {
	return &OfferHandler{db: db}
}

// CreateOffer creates a new offer on a listing (from the second provided offer.go)
func (h *OfferHandler) CreateOffer(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	listingIDStr := c.Param("listing_id") // Assuming listing_id is a URL parameter
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
			"error":   "Invalid request data: " + err.Error(),
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
	if req.ExpiresIn > 0 && req.ExpiresIn <= 168 { // Max 7 days (168 hours)
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

	offer, err := h.getOfferByID(offerID, userID, true) // Pass true for canBeOfferCreator
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

// GetOffer returns a specific offer (from the second provided offer.go)
func (h *OfferHandler) GetOffer(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
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

	offer, err := h.getOfferByID(offerID, userID, false) // User could be owner or creator
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Offer not found or not authorized",
		})
		return
	}

	// Mark as viewed if user is the listing owner and hasn't viewed it
	var listingOwnerID int
	var viewedAt sql.NullTime // Use sql.NullTime for nullable timestamp
	err = h.db.QueryRow(`
		SELECT l.user_id, o.viewed_at FROM offers o
		JOIN listings l ON o.listing_id = l.id
		WHERE o.id = $1
	`, offerID).Scan(&listingOwnerID, &viewedAt)

	if err == nil && listingOwnerID == userID && !viewedAt.Valid { // Check if viewed_at IS NULL
		_, dbErr := h.db.Exec(`
			UPDATE offers SET viewed_at = NOW(), updated_at = NOW() 
			WHERE id = $1 AND viewed_at IS NULL
		`, offerID)
		if dbErr == nil { // If update successful, re-fetch to get updated viewed_at
			updatedOffer, _ := h.getOfferByID(offerID, userID, false)
			if updatedOffer != nil {
				offer = updatedOffer
			}
		}
	}


	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    offer,
	})
}

// UpdateOffer updates an offer (accept, reject, counter-offer) (from the second provided offer.go)
func (h *OfferHandler) UpdateOffer(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
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

	if currentStatus != "pending" && currentStatus != "countered" { // Allow updates if pending or countered
		// Allow withdrawal by creator even if listing is closed by another offer
		if !(req.Status != nil && *req.Status == "withdrawn" && userID == fromUserID) {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Offer is not in a state that allows this update (" + currentStatus + ")",
			})
			return
		}
	}


	// Authorization based on action
	isListingOwner := (userID == listingOwnerID)
	isOfferCreator := (userID == fromUserID)

	if req.Status != nil {
		switch *req.Status {
		case "accepted", "rejected":
			if !isListingOwner {
				c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "Only listing owner can accept or reject offers"})
				return
			}
			if currentStatus != "pending" && currentStatus != "countered" { // Can only accept/reject pending/countered offers
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Offer must be pending or countered to be accepted/rejected"})
				return
			}
		case "withdrawn":
			if !isOfferCreator {
				c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "Only offer creator can withdraw offers"})
				return
			}
		case "countered": // Listing owner can counter
			if !isListingOwner {
				c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "Only listing owner can counter an offer"})
				return
			}
			if req.CounterOffer == nil || strings.TrimSpace(*req.CounterOffer) == "" {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Counter offer message cannot be empty"})
                return
			}
		default:
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid status update"})
			return
		}
	} else if req.CounterOffer != nil { // If only counter_offer is being sent, it implies a counter action by owner
		if !isListingOwner {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "Only listing owner can make a counter offer"})
			return
		}
		if currentStatus != "pending" { // Can only counter a pending offer
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Can only counter a pending offer"})
			return
		}
		// Implicitly set status to "countered" if CounterOffer is provided and status is not.
		statusCountered := "countered"
		req.Status = &statusCountered
	}


	if req.Status != nil && *req.Status == "accepted" {
		tx, err := h.db.Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to start transaction"})
			return
		}
		defer tx.Rollback()

		_, err = tx.Exec(`UPDATE offers SET status = 'accepted', updated_at = NOW() WHERE id = $1`, offerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to accept offer"})
			return
		}

		_, err = tx.Exec(`UPDATE listings SET status = 'closed', updated_at = NOW() WHERE id = $1`, listingID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to close listing"})
			return
		}

		_, err = tx.Exec(`UPDATE offers SET status = 'rejected', updated_at = NOW() 
			WHERE listing_id = $1 AND id != $2 AND (status = 'pending' OR status = 'countered')`, listingID, offerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to reject other offers"})
			return
		}

		if err = tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to commit transaction"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Offer accepted successfully"})
		return
	}

	// Handle other updates (reject, withdraw, counter-offer message, etc.)
	setParts := []string{}
	args := []interface{}{}
	argIdx := 1

	if req.Status != nil {
		setParts = append(setParts, "status = $"+strconv.Itoa(argIdx))
		args = append(args, *req.Status)
		argIdx++
	}
	if req.CounterOffer != nil {
		setParts = append(setParts, "counter_offer = $"+strconv.Itoa(argIdx))
		args = append(args, *req.CounterOffer)
		argIdx++
	}
	if req.Message != nil { // Allow updating message for certain scenarios (e.g. when countering)
		setParts = append(setParts, "message = $"+strconv.Itoa(argIdx))
		args = append(args, *req.Message)
		argIdx++
	}


	if len(setParts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "No fields to update"})
		return
	}

	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, offerID)
	query := `UPDATE offers SET ` + strings.Join(setParts, ", ") + ` WHERE id = $` + strconv.Itoa(argIdx)

	_, err = h.db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to update offer"})
		return
	}
	
	updatedOffer, err := h.getOfferByID(offerID, userID, isOfferCreator)
    if err != nil {
        c.JSON(http.StatusOK, gin.H{
            "success": true,
            "message": "Offer updated successfully, but failed to retrieve updated data",
        })
        return
    }

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Offer updated successfully",
		"data": updatedOffer,
	})
}


// GetUserOffers returns offers made by or received by the user (from the second provided offer.go)
func (h *OfferHandler) GetUserOffers(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "User not authenticated"})
		return
	}

	offerType := c.DefaultQuery("type", "all") // "sent", "received", "all"
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 { page = 1 }
	if limit < 1 || limit > 100 { limit = 20 }
	offset := (page - 1) * limit

	baseQuery := `
		SELECT o.id, o.listing_id, o.from_user_id, u_from.username AS from_username, u_from.avatar AS from_user_avatar,
		       o.proposed_product_id, p.name AS proposed_product_name, o.message, o.status, o.counter_offer,
		       o.expires_at, o.created_at, o.updated_at, o.viewed_at,
		       l.description AS listing_title 
		FROM offers o
		JOIN users u_from ON o.from_user_id = u_from.id
		JOIN user_products up ON o.proposed_product_id = up.id
		JOIN products p ON up.product_id = p.id
		JOIN listings l ON o.listing_id = l.id
	`
	countBaseQuery := `SELECT COUNT(o.id) FROM offers o JOIN listings l ON o.listing_id = l.id`

	whereConditions := []string{}
	queryArgs := []interface{}{}

	switch offerType {
	case "sent":
		whereConditions = append(whereConditions, "o.from_user_id = $"+strconv.Itoa(len(queryArgs)+1))
		queryArgs = append(queryArgs, userID)
	case "received":
		whereConditions = append(whereConditions, "l.user_id = $"+strconv.Itoa(len(queryArgs)+1))
		queryArgs = append(queryArgs, userID)
	default: // "all"
		whereConditions = append(whereConditions, "(o.from_user_id = $"+strconv.Itoa(len(queryArgs)+1)+" OR l.user_id = $"+strconv.Itoa(len(queryArgs)+2)+")")
		queryArgs = append(queryArgs, userID, userID)
	}

	if status != "" {
		whereConditions = append(whereConditions, "o.status = $"+strconv.Itoa(len(queryArgs)+1))
		queryArgs = append(queryArgs, status)
	}
	
	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Get total count
	var total int
	err := h.db.QueryRow(countBaseQuery+" "+whereClause, queryArgs...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to count offers: " + err.Error()})
		return
	}
	
	// Add limit and offset for data query
	limitOffsetArgs := make([]interface{}, len(queryArgs))
    copy(limitOffsetArgs, queryArgs)
	limitOffsetArgs = append(limitOffsetArgs, limit, offset)

	fullQuery := baseQuery + " " + whereClause + " ORDER BY o.created_at DESC LIMIT $" + strconv.Itoa(len(limitOffsetArgs)-1) + " OFFSET $" + strconv.Itoa(len(limitOffsetArgs))
	
	rows, err := h.db.Query(fullQuery, limitOffsetArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to retrieve offers: " + err.Error()})
		return
	}
	defer rows.Close()

	offers := []OfferResponse{}
	for rows.Next() {
		var offer OfferResponse
		// Ensure all fields in OfferResponse that are in the SELECT are scanned
		err := rows.Scan(
			&offer.ID, &offer.ListingID, &offer.FromUserID, &offer.FromUsername, &offer.FromUserAvatar,
			&offer.ProposedProductID, &offer.ProposedProductName, &offer.Message, &offer.Status, &offer.CounterOffer,
			&offer.ExpiresAt, &offer.CreatedAt, &offer.UpdatedAt, &offer.ViewedAt,
			&offer.ListingTitle,
		)
		if err != nil {
			// Log error
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

// GetOffersForListing retrieves all offers for a specific listing (listing owner only)
// Adapted from GetListingOffers in the original listing_offer.go
func (h *OfferHandler) GetOffersForListing(c *gin.Context) {
	currentUserID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	listingIDStr := c.Param("listing_id") // Renamed param for consistency
	listingID, err := strconv.Atoi(listingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid listing ID",
		})
		return
	}

	// Verify user owns the listing
	var ownerID int
	err = h.db.QueryRow("SELECT user_id FROM listings WHERE id = $1", listingID).Scan(&ownerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Listing not found",
		})
		return
	}

	if ownerID != currentUserID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Not authorized to view offers for this listing",
		})
		return
	}

	// Get offers
	rows, err := h.db.Query(`
		SELECT o.id, o.listing_id, o.from_user_id, u.username AS from_username, u.avatar AS from_user_avatar,
		       o.proposed_product_id, p.name AS proposed_product_name, o.message, o.status, 
		       o.counter_offer, o.expires_at, o.created_at, o.updated_at, o.viewed_at,
		       l.description AS listing_title 
		FROM offers o
		JOIN users u ON o.from_user_id = u.id
		JOIN user_products up ON o.proposed_product_id = up.id
		JOIN products p ON up.product_id = p.id
		JOIN listings l ON o.listing_id = l.id
		WHERE o.listing_id = $1
		ORDER BY o.created_at DESC
	`, listingID)

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
			&offer.ID, &offer.ListingID, &offer.FromUserID, &offer.FromUsername, &offer.FromUserAvatar,
			&offer.ProposedProductID, &offer.ProposedProductName, &offer.Message, &offer.Status,
			&offer.CounterOffer, &offer.ExpiresAt, &offer.CreatedAt, &offer.UpdatedAt, &offer.ViewedAt,
			&offer.ListingTitle,
			// Note: ListingDescription and ResponseDeadline are not in this query
		)
		if err != nil {
			// Log error or c.Error(err)
			continue
		}
		offers = append(offers, offer)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    offers,
	})
}


// GetOfferStats returns offer statistics for the user (from the second provided offer.go)
func (h *OfferHandler) GetOfferStats(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "User not authenticated"})
		return
	}

	var stats OfferStats
	stats.OffersByStatus = make(map[string]int) // Initialize map

	// Get total offers (sent or received)
	err := h.db.QueryRow(`
		SELECT COUNT(o.id) FROM offers o
		LEFT JOIN listings l ON o.listing_id = l.id
		WHERE o.from_user_id = $1 OR l.user_id = $1
	`, userID).Scan(&stats.TotalOffers)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to get total offers count"})
		return
	}

	// Get offers by status (sent or received)
	statusRows, err := h.db.Query(`
		SELECT o.status, COUNT(o.id)
		FROM offers o
		LEFT JOIN listings l ON o.listing_id = l.id
		WHERE o.from_user_id = $1 OR l.user_id = $1
		GROUP BY o.status
	`, userID)
	if err == nil {
		defer statusRows.Close()
		for statusRows.Next() {
			var status string
			var count int
			if scanErr := statusRows.Scan(&status, &count); scanErr == nil {
				stats.OffersByStatus[status] = count
			}
		}
	} else {
         c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to get offers by status"})
		return
    }


	// Calculate acceptance rate (for offers received by the user)
	var acceptedReceivedCount, totalReceivedForRateCount int
	err = h.db.QueryRow(`
		SELECT COUNT(CASE WHEN o.status = 'accepted' THEN 1 END),
		       COUNT(CASE WHEN o.status IN ('accepted', 'rejected', 'withdrawn', 'expired') THEN 1 END) -- Denominator: terminal states for received offers
		FROM offers o
		JOIN listings l ON o.listing_id = l.id
		WHERE l.user_id = $1
	`, userID).Scan(&acceptedReceivedCount, &totalReceivedForRateCount)
	if err == nil {
		if totalReceivedForRateCount > 0 {
			stats.AcceptanceRate = (float64(acceptedReceivedCount) / float64(totalReceivedForRateCount)) * 100
		} else {
			stats.AcceptanceRate = 0 // Avoid division by zero
		}
	} else {
        c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to calculate acceptance rate"})
		return
    }
    
    // AverageResponseTime and RecentOffers are not implemented in the original logic, so they remain zero/nil.
    // To implement AverageResponseTime, you'd need to store timestamps of offer creation and response (acceptance/rejection).
    // To implement RecentOffers, you would fetch a few recent offers similar to GetUserOffers but limited.

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// Helper function to get offer by ID with access control (from the second provided offer.go, slightly adapted)
// canBeOfferCreator helps in scenarios like CreateOffer where the querier is known to be the creator.
func (h *OfferHandler) getOfferByID(offerID, currentUserID int, isOfferCreator bool) (*OfferResponse, error) {
	var offer OfferResponse
	
	// Base query parts
    selectClause := `
        SELECT o.id, o.listing_id, o.from_user_id, u_from.username AS from_username, u_from.avatar AS from_user_avatar,
               o.proposed_product_id, p.name AS proposed_product_name, o.message, o.status, o.counter_offer,
               o.expires_at, o.created_at, o.updated_at, o.viewed_at,
               l.description AS listing_title, l.user_id AS listing_owner_id 
    `
    fromClause := `
        FROM offers o
        JOIN users u_from ON o.from_user_id = u_from.id
        JOIN user_products up ON o.proposed_product_id = up.id
        JOIN products p ON up.product_id = p.id
        JOIN listings l ON o.listing_id = l.id
    `
    // Dynamic WHERE clause based on whether the user is the offer creator or potentially the listing owner
    // This ensures that either the offer creator or the listing owner can fetch the offer.
    whereClause := `WHERE o.id = $1 AND (o.from_user_id = $2 OR l.user_id = $2)`
    if isOfferCreator { // If we know the user is the creator (e.g., right after creating an offer)
        whereClause = `WHERE o.id = $1 AND o.from_user_id = $2`
    }

	query := selectClause + fromClause + whereClause

    var listingOwnerID int // To capture listing_owner_id for authorization check inside the helper if needed

	err := h.db.QueryRow(query, offerID, currentUserID).Scan(
		&offer.ID, &offer.ListingID, &offer.FromUserID, &offer.FromUsername, &offer.FromUserAvatar,
		&offer.ProposedProductID, &offer.ProposedProductName, &offer.Message, &offer.Status, &offer.CounterOffer,
		&offer.ExpiresAt, &offer.CreatedAt, &offer.UpdatedAt, &offer.ViewedAt,
		&offer.ListingTitle, &listingOwnerID,
		// Note: ListingDescription and ResponseDeadline are not in this query and scan
	)

	if err != nil {
		return nil, err
	}
    
    // Additional authorization if not covered by the WHERE clause logic for specific cases (though current WHERE is broad enough)
    // if !isOfferCreator && offer.FromUserID != currentUserID && listingOwnerID != currentUserID {
	// 	 return nil, errors.New("user not authorized to view this offer")
	// }

	return &offer, nil
}