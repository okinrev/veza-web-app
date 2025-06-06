// internal/handlers/product.go
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

type ProductHandler struct {
	db *database.DB
}

type ProductResponse struct {
	ID                    int     `json:"id"`
	Name                  string  `json:"name"`
	CategoryID            *int    `json:"category_id"`
	CategoryName          string  `json:"category_name,omitempty"`
	Brand                 *string `json:"brand"`
	Model                 *string `json:"model"`
	Description           *string `json:"description"`
	Price                 *int    `json:"price"`
	WarrantyMonths        *int    `json:"warranty_months"`
	WarrantyConditions    *string `json:"warranty_conditions"`
	ManufacturerWebsite   *string `json:"manufacturer_website"`
	Specifications        *string `json:"specifications"`
	Status                string  `json:"status"`
	DocumentationCount    int     `json:"documentation_count"`
	UserCount             int     `json:"user_count"`
	CreatedAt             string  `json:"created_at"`
	UpdatedAt             string  `json:"updated_at"`
}

type CreateProductRequest struct {
	Name                string  `json:"name" binding:"required,min=1,max=200"`
	CategoryID          *int    `json:"category_id"`
	Brand               *string `json:"brand"`
	Model               *string `json:"model"`
	Description         *string `json:"description"`
	Price               *int    `json:"price"`
	WarrantyMonths      *int    `json:"warranty_months"`
	WarrantyConditions  *string `json:"warranty_conditions"`
	ManufacturerWebsite *string `json:"manufacturer_website"`
	Specifications      *string `json:"specifications"`
	Status              string  `json:"status"`
}

type UpdateProductRequest struct {
	Name                *string `json:"name,omitempty"`
	CategoryID          *int    `json:"category_id,omitempty"`
	Brand               *string `json:"brand,omitempty"`
	Model               *string `json:"model,omitempty"`
	Description         *string `json:"description,omitempty"`
	Price               *int    `json:"price,omitempty"`
	WarrantyMonths      *int    `json:"warranty_months,omitempty"`
	WarrantyConditions  *string `json:"warranty_conditions,omitempty"`
	ManufacturerWebsite *string `json:"manufacturer_website,omitempty"`
	Specifications      *string `json:"specifications,omitempty"`
	Status              *string `json:"status,omitempty"`
}

func NewProductHandler(db *database.DB) *ProductHandler {
	return &ProductHandler{db: db}
}

// GetProducts returns paginated list of products (admin only)
func (h *ProductHandler) GetProducts(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	// Verify admin access
	if !h.isAdmin(userID) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Admin access required",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := strings.TrimSpace(c.Query("search"))
	category := c.Query("category")
	status := c.Query("status")
	brand := c.Query("brand")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Build query
	baseQuery := `
		SELECT p.id, p.name, p.category_id, p.brand, p.model, p.description, 
		       p.price, p.warranty_months, p.warranty_conditions, p.manufacturer_website,
		       p.specifications, p.status, p.created_at, p.updated_at,
		       COALESCE(c.name, '') as category_name,
		       COALESCE(doc_count.count, 0) as documentation_count,
		       COALESCE(user_count.count, 0) as user_count
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		LEFT JOIN (
			SELECT product_id, COUNT(*) as count 
			FROM product_documents 
			GROUP BY product_id
		) doc_count ON p.id = doc_count.product_id
		LEFT JOIN (
			SELECT product_id, COUNT(*) as count 
			FROM user_products 
			GROUP BY product_id
		) user_count ON p.id = user_count.product_id
	`

	countQuery := "SELECT COUNT(*) FROM products p"
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argCount := 1

	if search != "" {
		whereClause += " AND (LOWER(p.name) LIKE $" + strconv.Itoa(argCount) + " OR LOWER(p.brand) LIKE $" + strconv.Itoa(argCount) + " OR LOWER(p.model) LIKE $" + strconv.Itoa(argCount) + ")"
		args = append(args, "%"+strings.ToLower(search)+"%")
		argCount++
	}

	if category != "" {
		categoryID, err := strconv.Atoi(category)
		if err == nil {
			whereClause += " AND p.category_id = $" + strconv.Itoa(argCount)
			args = append(args, categoryID)
			argCount++
		}
	}

	if status != "" {
		whereClause += " AND p.status = $" + strconv.Itoa(argCount)
		args = append(args, status)
		argCount++
	}

	if brand != "" {
		whereClause += " AND LOWER(p.brand) = LOWER($" + strconv.Itoa(argCount) + ")"
		args = append(args, brand)
		argCount++
	}

	// Get total count
	var total int
	err := h.db.QueryRow(countQuery+" "+whereClause, args...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to count products",
		})
		return
	}

	// Get products
	orderClause := " ORDER BY p.updated_at DESC LIMIT $" + strconv.Itoa(argCount) + " OFFSET $" + strconv.Itoa(argCount+1)
	args = append(args, limit, offset)

	rows, err := h.db.Query(baseQuery+" "+whereClause+orderClause, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve products",
		})
		return
	}
	defer rows.Close()

	products := []ProductResponse{}
	for rows.Next() {
		var product ProductResponse
		err := rows.Scan(
			&product.ID, &product.Name, &product.CategoryID, &product.Brand, &product.Model,
			&product.Description, &product.Price, &product.WarrantyMonths, &product.WarrantyConditions,
			&product.ManufacturerWebsite, &product.Specifications, &product.Status, &product.CreatedAt,
			&product.UpdatedAt, &product.CategoryName, &product.DocumentationCount, &product.UserCount,
		)
		if err != nil {
			continue
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

// GetProduct returns a specific product by ID
func (h *ProductHandler) GetProduct(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	if !h.isAdmin(userID) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Admin access required",
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

	product, err := h.getProductByID(productID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Product not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    product,
	})
}

// CreateProduct creates a new product (admin only)
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	if !h.isAdmin(userID) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Admin access required",
		})
		return
	}

	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
		})
		return
	}

	// Set default status if not provided
	if req.Status == "" {
		req.Status = "active"
	}

	// Create product
	var productID int
	err := h.db.QueryRow(`
		INSERT INTO products (
			name, category_id, brand, model, description, price, 
			warranty_months, warranty_conditions, manufacturer_website, 
			specifications, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW()) 
		RETURNING id
	`, req.Name, req.CategoryID, req.Brand, req.Model, req.Description, req.Price,
		req.WarrantyMonths, req.WarrantyConditions, req.ManufacturerWebsite,
		req.Specifications, req.Status).Scan(&productID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create product",
		})
		return
	}

	// Return created product
	product, err := h.getProductByID(productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Product created but failed to retrieve data",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Product created successfully",
		"data":    product,
	})
}

// UpdateProduct updates a product (admin only)
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	if !h.isAdmin(userID) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Admin access required",
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

	var req UpdateProductRequest
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

	if req.Name != nil {
		setParts = append(setParts, "name = $"+strconv.Itoa(argCount))
		args = append(args, strings.TrimSpace(*req.Name))
		argCount++
	}
	if req.CategoryID != nil {
		setParts = append(setParts, "category_id = $"+strconv.Itoa(argCount))
		args = append(args, req.CategoryID)
		argCount++
	}
	if req.Brand != nil {
		setParts = append(setParts, "brand = $"+strconv.Itoa(argCount))
		args = append(args, req.Brand)
		argCount++
	}
	if req.Model != nil {
		setParts = append(setParts, "model = $"+strconv.Itoa(argCount))
		args = append(args, req.Model)
		argCount++
	}
	if req.Description != nil {
		setParts = append(setParts, "description = $"+strconv.Itoa(argCount))
		args = append(args, req.Description)
		argCount++
	}
	if req.Price != nil {
		setParts = append(setParts, "price = $"+strconv.Itoa(argCount))
		args = append(args, req.Price)
		argCount++
	}
	if req.WarrantyMonths != nil {
		setParts = append(setParts, "warranty_months = $"+strconv.Itoa(argCount))
		args = append(args, req.WarrantyMonths)
		argCount++
	}
	if req.WarrantyConditions != nil {
		setParts = append(setParts, "warranty_conditions = $"+strconv.Itoa(argCount))
		args = append(args, req.WarrantyConditions)
		argCount++
	}
	if req.ManufacturerWebsite != nil {
		setParts = append(setParts, "manufacturer_website = $"+strconv.Itoa(argCount))
		args = append(args, req.ManufacturerWebsite)
		argCount++
	}
	if req.Specifications != nil {
		setParts = append(setParts, "specifications = $"+strconv.Itoa(argCount))
		args = append(args, req.Specifications)
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

	// Add updated_at and product_id
	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, productID)

	query := "UPDATE products SET " + strings.Join(setParts, ", ") + " WHERE id = $" + strconv.Itoa(argCount)

	_, err = h.db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update product",
		})
		return
	}

	// Return updated product
	product, err := h.getProductByID(productID)
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

// DeleteProduct deletes a product (admin only)
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	if !h.isAdmin(userID) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Admin access required",
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

	// Check if product is being used
	var count int
	err = h.db.QueryRow("SELECT COUNT(*) FROM user_products WHERE product_id = $1", productID).Scan(&count)
	if err == nil && count > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"error":   "Cannot delete product: it is being used by users",
		})
		return
	}

	// Delete product
	_, err = h.db.Exec("DELETE FROM products WHERE id = $1", productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to delete product",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Product deleted successfully",
	})
}

// SearchProducts allows users to search products (public endpoint)
func (h *ProductHandler) SearchProducts(c *gin.Context) {
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

	// Search only active products
	rows, err := h.db.Query(`
		SELECT p.id, p.name, p.category_id, p.brand, p.model, p.description, 
		       p.price, p.warranty_months, p.warranty_conditions, p.manufacturer_website,
		       p.specifications, p.status, p.created_at, p.updated_at,
		       COALESCE(c.name, '') as category_name
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.status = 'active' AND (
			LOWER(p.name) LIKE $1 OR 
			LOWER(p.brand) LIKE $1 OR 
			LOWER(p.model) LIKE $1 OR 
			LOWER(p.description) LIKE $1
		)
		ORDER BY p.name ASC
		LIMIT $2
	`, "%"+strings.ToLower(query)+"%", limit)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to search products",
		})
		return
	}
	defer rows.Close()

	products := []ProductResponse{}
	for rows.Next() {
		var product ProductResponse
		err := rows.Scan(
			&product.ID, &product.Name, &product.CategoryID, &product.Brand, &product.Model,
			&product.Description, &product.Price, &product.WarrantyMonths, &product.WarrantyConditions,
			&product.ManufacturerWebsite, &product.Specifications, &product.Status, &product.CreatedAt,
			&product.UpdatedAt, &product.CategoryName,
		)
		if err != nil {
			continue
		}
		// Don't include counts for public search
		products = append(products, product)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    products,
	})
}

// Helper functions
func (h *ProductHandler) getProductByID(productID int) (*ProductResponse, error) {
	var product ProductResponse
	err := h.db.QueryRow(`
		SELECT p.id, p.name, p.category_id, p.brand, p.model, p.description, 
		       p.price, p.warranty_months, p.warranty_conditions, p.manufacturer_website,
		       p.specifications, p.status, p.created_at, p.updated_at,
		       COALESCE(c.name, '') as category_name,
		       COALESCE(doc_count.count, 0) as documentation_count,
		       COALESCE(user_count.count, 0) as user_count
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		LEFT JOIN (
			SELECT product_id, COUNT(*) as count 
			FROM product_documents 
			WHERE product_id = $1
			GROUP BY product_id
		) doc_count ON p.id = doc_count.product_id
		LEFT JOIN (
			SELECT product_id, COUNT(*) as count 
			FROM user_products 
			WHERE product_id = $1
			GROUP BY product_id
		) user_count ON p.id = user_count.product_id
		WHERE p.id = $1
	`, productID).Scan(
		&product.ID, &product.Name, &product.CategoryID, &product.Brand, &product.Model,
		&product.Description, &product.Price, &product.WarrantyMonths, &product.WarrantyConditions,
		&product.ManufacturerWebsite, &product.Specifications, &product.Status, &product.CreatedAt,
		&product.UpdatedAt, &product.CategoryName, &product.DocumentationCount, &product.UserCount,
	)
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (h *ProductHandler) isAdmin(userID int) bool {
	var role string
	err := h.db.QueryRow("SELECT role FROM users WHERE id = $1", userID).Scan(&role)
	if err != nil {
		return false
	}
	return role == "admin" || role == "super_admin"
}