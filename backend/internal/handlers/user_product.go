// internal/handlers/user_product.go
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"veza-web-app/internal/database"
	"veza-web-app/internal/middleware"
	"veza-web-app/internal/models"
)

type UserProductHandler struct {
	db *database.DB
}

type UserProductResponse struct {
	ID              int     `json:"id"`
	UserID          int     `json:"user_id"`
	ProductID       int     `json:"product_id"`
	ProductName     string  `json:"product_name"`
	CategoryName    string  `json:"category_name,omitempty"`
	Brand           *string `json:"brand,omitempty"`
	Model           *string `json:"model,omitempty"`
	Version         *string `json:"version"`
	PurchaseDate    *string `json:"purchase_date"`
	WarrantyExpires *string `json:"warranty_expires"`
	PurchasePrice   *int    `json:"purchase_price"`
	SerialNumber    *string `json:"serial_number"`
	Notes           *string `json:"notes"`
	Status          string  `json:"status"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
	
	// Additional computed fields
	FilesCount      int  `json:"files_count,omitempty"`
	DocsCount       int  `json:"docs_count,omitempty"`
	IsUnderWarranty bool `json:"is_under_warranty,omitempty"`
}

type CreateUserProductRequest struct {
	ProductID       int     `json:"product_id" binding:"required"`
	Version         *string `json:"version"`
	PurchaseDate    *string `json:"purchase_date"`
	WarrantyExpires *string `json:"warranty_expires"`
	PurchasePrice   *int    `json:"purchase_price"`
	SerialNumber    *string `json:"serial_number"`
	Notes           *string `json:"notes"`
}

type UpdateUserProductRequest struct {
	Version         *string `json:"version,omitempty"`
	PurchaseDate    *string `json:"purchase_date,omitempty"`
	WarrantyExpires *string `json:"warranty_expires,omitempty"`
	PurchasePrice   *int    `json:"purchase_price,omitempty"`
	SerialNumber    *string `json:"serial_number,omitempty"`
	Notes           *string `json:"notes,omitempty"`
	Status          *string `json:"status,omitempty"`
}

func NewUserProductHandler(db *database.DB) *UserProductHandler {
	return &UserProductHandler{db: db}
}

// ListUserProducts returns user's owned products
func (h *UserProductHandler) ListUserProducts(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.DefaultQuery("status", "")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Build query
	baseQuery := `
		SELECT up.id, up.user_id, up.product_id, p.name as product_name, 
		       COALESCE(c.name, '') as category_name, p.brand, p.model,
		       up.version, up.purchase_date, up.warranty_expires, up.purchase_price,
		       up.serial_number, up.notes, up.status, up.created_at, up.updated_at,
		       COALESCE(f.files_count, 0) as files_count,
		       COALESCE(d.docs_count, 0) as docs_count
		FROM user_products up
		JOIN products p ON up.product_id = p.id
		LEFT JOIN categories c ON p.category_id = c.id
		LEFT JOIN (
			SELECT product_id, COUNT(*) as files_count 
			FROM files 
			GROUP BY product_id
		) f ON up.id = f.product_id
		LEFT JOIN (
			SELECT product_id, COUNT(*) as docs_count 
			FROM internal_documents 
			GROUP BY product_id
		) d ON up.id = d.product_id
		WHERE up.user_id = $1
	`

	countQuery := "SELECT COUNT(*) FROM user_products WHERE user_id = $1"
	args := []interface{}{userID}
	argCount := 2

	if status != "" {
		baseQuery += " AND up.status = $" + strconv.Itoa(argCount)
		countQuery += " AND status = $" + strconv.Itoa(argCount)
		args = append(args, status)
		argCount++
	}

	// Get total count
	var total int
	err := h.db.QueryRow(countQuery, args[:len(args)-1]...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to add product to collection",
		})
		return
	}

	// Return created user product
	product, err := h.getUserProductByID(userProductID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Product added but failed to retrieve data",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Product added to collection successfully",
		"data":    product,
	})
}

// UpdateUserProduct updates a user's product
func (h *UserProductHandler) UpdateUserProduct(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	idStr := c.Param("id")
	userProductID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid product ID",
		})
		return
	}

	// Verify ownership
	var ownerID int
	err = h.db.QueryRow("SELECT user_id FROM user_products WHERE id = $1", userProductID).Scan(&ownerID)
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
			"error":   "Not authorized to update this product",
		})
		return
	}

	var req UpdateUserProductRequest
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

	if req.Version != nil {
		setParts = append(setParts, "version = $"+strconv.Itoa(argCount))
		args = append(args, req.Version)
		argCount++
	}
	if req.PurchaseDate != nil {
		setParts = append(setParts, "purchase_date = $"+strconv.Itoa(argCount))
		args = append(args, req.PurchaseDate)
		argCount++
	}
	if req.WarrantyExpires != nil {
		setParts = append(setParts, "warranty_expires = $"+strconv.Itoa(argCount))
		args = append(args, req.WarrantyExpires)
		argCount++
	}
	if req.PurchasePrice != nil {
		setParts = append(setParts, "purchase_price = $"+strconv.Itoa(argCount))
		args = append(args, req.PurchasePrice)
		argCount++
	}
	if req.SerialNumber != nil {
		setParts = append(setParts, "serial_number = $"+strconv.Itoa(argCount))
		args = append(args, req.SerialNumber)
		argCount++
	}
	if req.Notes != nil {
		setParts = append(setParts, "notes = $"+strconv.Itoa(argCount))
		args = append(args, req.Notes)
		argCount++
	}
	if req.Status != nil {
		setParts = append(setParts, "status = $"+strconv.Itoa(argCount))
		args = append(args, req.Status)
		argCount++
	}

	if len(setParts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "No fields to update",
		})
		return
	}

	// Add updated_at and user_product_id
	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, userProductID)

	query := "UPDATE user_products SET " + strings.Join(setParts, ", ") + " WHERE id = $" + strconv.Itoa(argCount)

	_, err = h.db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update product",
		})
		return
	}

	// Return updated product
	product, err := h.getUserProductByID(userProductID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Product updated but failed to retrieve updated data",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Product updated successfully",
		"data":    product,
	})
}

// DeleteUserProduct removes a product from user's collection
func (h *UserProductHandler) DeleteUserProduct(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	idStr := c.Param("id")
	userProductID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid product ID",
		})
		return
	}

	// Verify ownership
	var ownerID int
	err = h.db.QueryRow("SELECT user_id FROM user_products WHERE id = $1", userProductID).Scan(&ownerID)
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
			"error":   "Not authorized to delete this product",
		})
		return
	}

	// Check if there are related files or documents
	var filesCount, docsCount int
	h.db.QueryRow("SELECT COUNT(*) FROM files WHERE product_id = $1", userProductID).Scan(&filesCount)
	h.db.QueryRow("SELECT COUNT(*) FROM internal_documents WHERE product_id = $1", userProductID).Scan(&docsCount)

	if filesCount > 0 || docsCount > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"error":   "Cannot delete product: it has associated files or documents. Please remove them first.",
		})
		return
	}

	// Delete user product
	_, err = h.db.Exec("DELETE FROM user_products WHERE id = $1", userProductID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to delete product from collection",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Product removed from collection successfully",
	})
}

// GetWarrantyStatus returns warranty status for user's products
func (h *UserProductHandler) GetWarrantyStatus(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	filter := c.Query("filter") // "expiring", "expired", "active"

	baseQuery := `
		SELECT up.id, p.name as product_name, up.warranty_expires, up.purchase_date
		FROM user_products up
		JOIN products p ON up.product_id = p.id
		WHERE up.user_id = $1 AND up.warranty_expires IS NOT NULL
	`

	args := []interface{}{userID}

	switch filter {
	case "expiring":
		// Expiring within 30 days
		baseQuery += " AND up.warranty_expires BETWEEN NOW() AND NOW() + INTERVAL '30 days'"
	case "expired":
		// Already expired
		baseQuery += " AND up.warranty_expires < NOW()"
	case "active":
		// Still under warranty
		baseQuery += " AND up.warranty_expires >= NOW()"
	}

	baseQuery += " ORDER BY up.warranty_expires ASC"

	rows, err := h.db.Query(baseQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve warranty information",
		})
		return
	}
	defer rows.Close()

	warranties := []map[string]interface{}{}
	for rows.Next() {
		var id int
		var productName string
		var warrantyExpires, purchaseDate *string

		err := rows.Scan(&id, &productName, &warrantyExpires, &purchaseDate)
		if err != nil {
			continue
		}

		warranty := map[string]interface{}{
			"id":               id,
			"product_name":     productName,
			"warranty_expires": warrantyExpires,
			"purchase_date":    purchaseDate,
		}

		if warrantyExpires != nil {
			warrantyDate, err := time.Parse("2006-01-02", *warrantyExpires)
			if err == nil {
				warranty["is_under_warranty"] = warrantyDate.After(time.Now())
				warranty["days_remaining"] = int(time.Until(warrantyDate).Hours() / 24)
			}
		}

		warranties = append(warranties, warranty)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    warranties,
	})
}

// SearchUserProducts searches within user's products
func (h *UserProductHandler) SearchUserProducts(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Query parameter 'q' is required",
		})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	rows, err := h.db.Query(`
		SELECT up.id, up.user_id, up.product_id, p.name as product_name, 
		       COALESCE(c.name, '') as category_name, p.brand, p.model,
		       up.version, up.purchase_date, up.warranty_expires, up.purchase_price,
		       up.serial_number, up.notes, up.status, up.created_at, up.updated_at
		FROM user_products up
		JOIN products p ON up.product_id = p.id
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE up.user_id = $1 AND (
			LOWER(p.name) LIKE $2 OR 
			LOWER(p.brand) LIKE $2 OR 
			LOWER(p.model) LIKE $2 OR 
			LOWER(up.version) LIKE $2 OR
			LOWER(up.serial_number) LIKE $2 OR
			LOWER(up.notes) LIKE $2
		)
		ORDER BY p.name ASC
		LIMIT $3
	`, userID, "%"+strings.ToLower(query)+"%", limit)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to search products",
		})
		return
	}
	defer rows.Close()

	products := []UserProductResponse{}
	for rows.Next() {
		var product UserProductResponse
		err := rows.Scan(
			&product.ID, &product.UserID, &product.ProductID, &product.ProductName,
			&product.CategoryName, &product.Brand, &product.Model, &product.Version,
			&product.PurchaseDate, &product.WarrantyExpires, &product.PurchasePrice,
			&product.SerialNumber, &product.Notes, &product.Status, &product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			continue
		}

		// Check if under warranty
		if product.WarrantyExpires != nil {
			warrantyDate, err := time.Parse("2006-01-02", *product.WarrantyExpires)
			if err == nil {
				product.IsUnderWarranty = warrantyDate.After(time.Now())
			}
		}

		products = append(products, product)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    products,
	})
}

// Helper function to get user product by ID
func (h *UserProductHandler) getUserProductByID(userProductID, userID int) (*UserProductResponse, error) {
	var product UserProductResponse
	err := h.db.QueryRow(`
		SELECT up.id, up.user_id, up.product_id, p.name as product_name, 
		       COALESCE(c.name, '') as category_name, p.brand, p.model,
		       up.version, up.purchase_date, up.warranty_expires, up.purchase_price,
		       up.serial_number, up.notes, up.status, up.created_at, up.updated_at,
		       COALESCE(f.files_count, 0) as files_count,
		       COALESCE(d.docs_count, 0) as docs_count
		FROM user_products up
		JOIN products p ON up.product_id = p.id
		LEFT JOIN categories c ON p.category_id = c.id
		LEFT JOIN (
			SELECT product_id, COUNT(*) as files_count 
			FROM files 
			WHERE product_id = $1
			GROUP BY product_id
		) f ON up.id = f.product_id
		LEFT JOIN (
			SELECT product_id, COUNT(*) as docs_count 
			FROM internal_documents 
			WHERE product_id = $1
			GROUP BY product_id
		) d ON up.id = d.product_id
		WHERE up.id = $1 AND up.user_id = $2
	`, userProductID, userID).Scan(
		&product.ID, &product.UserID, &product.ProductID, &product.ProductName,
		&product.CategoryName, &product.Brand, &product.Model, &product.Version,
		&product.PurchaseDate, &product.WarrantyExpires, &product.PurchasePrice,
		&product.SerialNumber, &product.Notes, &product.Status, &product.CreatedAt,
		&product.UpdatedAt, &product.FilesCount, &product.DocsCount,
	)
	if err != nil {
		return nil, err
	}

	// Check if under warranty
	if product.WarrantyExpires != nil {
		warrantyDate, err := time.Parse("2006-01-02", *product.WarrantyExpires)
		if err == nil {
			product.IsUnderWarranty = warrantyDate.After(time.Now())
		}
	}

	return &product, nil
} {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to count user products",
		})
		return
	}

	// Get products
	orderClause := " ORDER BY up.purchase_date DESC NULLS LAST, up.created_at DESC LIMIT $" + strconv.Itoa(argCount) + " OFFSET $" + strconv.Itoa(argCount+1)
	args = append(args, limit, offset)

	rows, err := h.db.Query(baseQuery+orderClause, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve user products",
		})
		return
	}
	defer rows.Close()

	products := []UserProductResponse{}
	for rows.Next() {
		var product UserProductResponse
		err := rows.Scan(
			&product.ID, &product.UserID, &product.ProductID, &product.ProductName,
			&product.CategoryName, &product.Brand, &product.Model, &product.Version,
			&product.PurchaseDate, &product.WarrantyExpires, &product.PurchasePrice,
			&product.SerialNumber, &product.Notes, &product.Status, &product.CreatedAt,
			&product.UpdatedAt, &product.FilesCount, &product.DocsCount,
		)
		if err != nil {
			continue
		}

		// Check if under warranty
		if product.WarrantyExpires != nil {
			warrantyDate, err := time.Parse("2006-01-02", *product.WarrantyExpires)
			if err == nil {
				product.IsUnderWarranty = warrantyDate.After(time.Now())
			}
		}

		products = append(products, product)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    products,
		"meta": gin.H{
			"page":        page,
			"per_page":    limit,
			"total":       total,
			"total_pages": (total + limit - 1) / limit,
		},
	})
}

// GetUserProduct returns a specific user product
func (h *UserProductHandler) GetUserProduct(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	idStr := c.Param("id")
	productID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid product ID",
		})
		return
	}

	product, err := h.getUserProductByID(productID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Product not found or not owned by user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    product,
	})
}

// CreateUserProduct adds a product to user's collection
func (h *UserProductHandler) CreateUserProduct(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	var req CreateUserProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
		})
		return
	}

	// Verify that the product exists in the catalog
	var productExists int
	err := h.db.QueryRow("SELECT COUNT(*) FROM products WHERE id = $1 AND status = 'active'", req.ProductID).Scan(&productExists)
	if err != nil || productExists == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Product not found in catalog or not active",
		})
		return
	}

	// Check if user already owns this product
	var existingCount int
	err = h.db.QueryRow("SELECT COUNT(*) FROM user_products WHERE user_id = $1 AND product_id = $2", userID, req.ProductID).Scan(&existingCount)
	if err == nil && existingCount > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"error":   "Product already exists in your collection",
		})
		return
	}

	// Create user product
	var userProductID int
	err = h.db.QueryRow(`
		INSERT INTO user_products (user_id, product_id, version, purchase_date, warranty_expires, 
		                           purchase_price, serial_number, notes, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'active', NOW(), NOW())
		RETURNING id
	`, userID, req.ProductID, req.Version, req.PurchaseDate, req.WarrantyExpires,
		req.PurchasePrice, req.SerialNumber, req.Notes).Scan(&userProductID)

	if err != nil