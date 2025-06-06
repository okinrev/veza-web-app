// internal/handlers/listing.go
package handlers

import (
	"veza-web-app/internal/common"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"veza-web-app/internal/database"
	"veza-web-app/internal/middleware"
	// "veza-web-app/internal/models" // This import was in the original file but not used by the moved code.
)

type ListingHandler struct {
	db *database.DB
}

type ListingResponse struct {
	ID          int      `json:"id"`
	UserID      int      `json:"user_id"`
	Username    string   `json:"username,omitempty"`
	ProductID   int      `json:"product_id"` // This is user_products.id
	ProductName string   `json:"product_name,omitempty"`
	Description string   `json:"description"`
	State       string   `json:"state"`
	Price       *int     `json:"price"`
	ExchangeFor *string  `json:"exchange_for"`
	Images      []string `json:"images"`
	Status      string   `json:"status"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

type CreateListingRequest struct {
	ProductID   int      `json:"product_id" binding:"required"` // This is user_products.id
	Description string   `json:"description" binding:"required"`
	State       string   `json:"state" binding:"required"`
	Price       *int     `json:"price"`
	ExchangeFor *string  `json:"exchange_for"`
	Images      []string `json:"images"`
}

type UpdateListingRequest struct {
	Description *string  `json:"description,omitempty"`
	State       *string  `json:"state,omitempty"`
	Price       *int     `json:"price,omitempty"`
	ExchangeFor *string  `json:"exchange_for,omitempty"`
	Images      *[]string `json:"images,omitempty"`
	Status      *string  `json:"status,omitempty"`
}

func NewListingHandler(db *database.DB) *ListingHandler {
	return &ListingHandler{db: db}
}

// CreateListing creates a new listing
func (h *ListingHandler) CreateListing(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	var req CreateListingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
		})
		return
	}

	// Verify user owns the product (user_products.id)
	var productOwnerID int
	//var productName string // productName is fetched later when retrieving the full listing
	err := h.db.QueryRow("SELECT user_id FROM user_products WHERE id = $1", req.ProductID).Scan(&productOwnerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Product not found or not owned by user",
		})
		return
	}

	if productOwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Not authorized to create listing for this product",
		})
		return
	}

	// Create listing
	var listingID int
	err = h.db.QueryRow(`
		INSERT INTO listings (user_id, product_id, description, state, price, exchange_for, images, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'open', NOW(), NOW())
		RETURNING id
	`, userID, req.ProductID, req.Description, req.State, req.Price, req.ExchangeFor, pq.Array(req.Images)).Scan(&listingID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create listing",
		})
		return
	}

	// Return created listing
	listing, err := h.getListingByID(listingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Listing created but failed to retrieve data",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Listing created successfully",
		"data":    listing,
	})
}

// GetAllListings returns all active listings
func (h *ListingHandler) GetAllListings(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.DefaultQuery("status", "open") // Default to 'open' listings

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Get total count
	var total int
	err := h.db.QueryRow("SELECT COUNT(*) FROM listings WHERE status = $1", status).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to count listings",
		})
		return
	}

	// Get listings with user and product info
	rows, err := h.db.Query(`
		SELECT l.id, l.user_id, u.username, l.product_id, p.name, l.description, l.state, 
		       l.price, l.exchange_for, l.images, l.status, l.created_at, l.updated_at
		FROM listings l
		JOIN users u ON l.user_id = u.id
		JOIN user_products up ON l.product_id = up.id
		JOIN products p ON up.product_id = p.id
		WHERE l.status = $1
		ORDER BY l.created_at DESC
		LIMIT $2 OFFSET $3
	`, status, limit, offset)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve listings",
		})
		return
	}
	defer rows.Close()

	listings := []ListingResponse{}
	for rows.Next() {
		var listing ListingResponse
		var images pq.StringArray
		err := rows.Scan(
			&listing.ID, &listing.UserID, &listing.Username, &listing.ProductID,
			&listing.ProductName, &listing.Description, &listing.State, &listing.Price,
			&listing.ExchangeFor, &images, &listing.Status, &listing.CreatedAt, &listing.UpdatedAt,
		)
		if err != nil {
			// Log error: c.Error(err) or log.Println(err)
			continue
		}
		listing.Images = []string(images)
		listings = append(listings, listing)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    listings,
		"meta": gin.H{
			"page":        page,
			"per_page":    limit,
			"total":       total,
			"total_pages": (total + limit - 1) / limit,
		},
	})
}

// GetListingByID returns a specific listing
func (h *ListingHandler) GetListingByID(c *gin.Context) {
	idStr := c.Param("id")
	listingID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid listing ID",
		})
		return
	}

	listing, err := h.getListingByID(listingID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Listing not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    listing,
	})
}

// UpdateListing updates a listing (owner only)
func (h *ListingHandler) UpdateListing(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	idStr := c.Param("id")
	listingID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid listing ID",
		})
		return
	}

	// Verify ownership
	var ownerID int
	err = h.db.QueryRow("SELECT user_id FROM listings WHERE id = $1", listingID).Scan(&ownerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Listing not found",
		})
		return
	}

	if ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Not authorized to update this listing",
		})
		return
	}

	var req UpdateListingRequest
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

	if req.Description != nil {
		setParts = append(setParts, "description = $"+strconv.Itoa(argCount))
		args = append(args, strings.TrimSpace(*req.Description))
		argCount++
	}
	if req.State != nil {
		setParts = append(setParts, "state = $"+strconv.Itoa(argCount))
		args = append(args, *req.State)
		argCount++
	}
	if req.Price != nil {
		setParts = append(setParts, "price = $"+strconv.Itoa(argCount))
		args = append(args, req.Price)
		argCount++
	}
	if req.ExchangeFor != nil {
		setParts = append(setParts, "exchange_for = $"+strconv.Itoa(argCount))
		args = append(args, req.ExchangeFor)
		argCount++
	}
	if req.Images != nil {
		setParts = append(setParts, "images = $"+strconv.Itoa(argCount))
		args = append(args, pq.Array(*req.Images))
		argCount++
	}
	if req.Status != nil {
		setParts = append(setParts, "status = $"+strconv.Itoa(argCount))
		args = append(args, *req.Status)
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
	args = append(args, listingID)

	query := "UPDATE listings SET " + strings.Join(setParts, ", ") + " WHERE id = $" + strconv.Itoa(argCount)

	_, err = h.db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update listing",
		})
		return
	}

	listing, err := h.getListingByID(listingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Listing updated but failed to retrieve updated data",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Listing updated successfully",
		"data":    listing,
	})
}

// DeleteListing deletes a listing (owner only)
func (h *ListingHandler) DeleteListing(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	idStr := c.Param("id")
	listingID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid listing ID",
		})
		return
	}

	// Verify ownership
	var ownerID int
	err = h.db.QueryRow("SELECT user_id FROM listings WHERE id = $1", listingID).Scan(&ownerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Listing not found",
		})
		return
	}

	if ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Not authorized to delete this listing",
		})
		return
	}

	// Delete listing (cascade should handle related offers if DB is set up for it)
	_, err = h.db.Exec("DELETE FROM listings WHERE id = $1", listingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to delete listing",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Listing deleted successfully",
	})
}

// Helper function
func (h *ListingHandler) getListingByID(listingID int) (*ListingResponse, error) {
	var listing ListingResponse
	var images pq.StringArray
	err := h.db.QueryRow(`
		SELECT l.id, l.user_id, u.username, l.product_id, p.name, l.description, l.state, 
		       l.price, l.exchange_for, l.images, l.status, l.created_at, l.updated_at
		FROM listings l
		JOIN users u ON l.user_id = u.id
		JOIN user_products up ON l.product_id = up.id
		JOIN products p ON up.product_id = p.id
		WHERE l.id = $1
	`, listingID).Scan(
		&listing.ID, &listing.UserID, &listing.Username, &listing.ProductID,
		&listing.ProductName, &listing.Description, &listing.State, &listing.Price,
		&listing.ExchangeFor, &images, &listing.Status, &listing.CreatedAt, &listing.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	listing.Images = []string(images)
	return &listing, nil
}