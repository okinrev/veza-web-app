//file: internal/models/admin/product.go

package admin

import (
	"time"
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// ProductSpecs représente les spécifications techniques en JSON
type ProductSpecs map[string]interface{}

func (ps ProductSpecs) Value() (driver.Value, error) {
	return json.Marshal(ps)
}

func (ps *ProductSpecs) Scan(value interface{}) error {
	if value == nil {
		*ps = make(ProductSpecs)
		return nil
	}
	
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, ps)
	case string:
		return json.Unmarshal([]byte(v), ps)
	default:
		return errors.New("cannot scan ProductSpecs")
	}
}

// Product modèle étendu pour l'administration
type Product struct {
	ID                   int64         `json:"id" db:"id"`
	Name                 string        `json:"name" db:"name" validate:"required,min=2,max=255"`
	CategoryID           *int64        `json:"category_id" db:"category_id"`
	Brand                string        `json:"brand" db:"brand" validate:"max=100"`
	Model                string        `json:"model" db:"model" validate:"max=100"`
	Description          string        `json:"description" db:"description" validate:"max=2000"`
	Price                *float64      `json:"price" db:"price" validate:"omitempty,min=0"`
	WarrantyMonths       int           `json:"warranty_months" db:"warranty_months" validate:"min=0,max=360"`
	WarrantyConditions   string        `json:"warranty_conditions" db:"warranty_conditions" validate:"max=1000"`
	ManufacturerWebsite  string        `json:"manufacturer_website" db:"manufacturer_website" validate:"omitempty,url"`
	Specifications       ProductSpecs  `json:"specifications" db:"specifications"`
	Status               string        `json:"status" db:"status" validate:"required,oneof=active discontinued draft"`
	CreatedAt            time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time     `json:"updated_at" db:"updated_at"`
	
	// Champs calculés/jointures
	CategoryName         string        `json:"category_name,omitempty" db:"category_name"`
	DocumentationCount   int           `json:"documentation_count,omitempty" db:"documentation_count"`
	UserCount           int           `json:"user_count,omitempty" db:"user_count"`
	TotalSales          *float64      `json:"total_sales,omitempty" db:"total_sales"`
}

// CreateProductRequest structure pour la création
type CreateProductRequest struct {
	Name                string       `json:"name" validate:"required,min=2,max=255"`
	CategoryID          *int64       `json:"category_id"`
	Brand               string       `json:"brand" validate:"max=100"`
	Model               string       `json:"model" validate:"max=100"`
	Description         string       `json:"description" validate:"max=2000"`
	Price               *float64     `json:"price" validate:"omitempty,min=0"`
	WarrantyMonths      int          `json:"warranty_months" validate:"min=0,max=360"`
	WarrantyConditions  string       `json:"warranty_conditions" validate:"max=1000"`
	ManufacturerWebsite string       `json:"manufacturer_website" validate:"omitempty,url"`
	Specifications      ProductSpecs `json:"specifications"`
	Status              string       `json:"status" validate:"required,oneof=active discontinued draft"`
}

// UpdateProductRequest structure pour la mise à jour
type UpdateProductRequest struct {
	Name                *string      `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	CategoryID          *int64       `json:"category_id,omitempty"`
	Brand               *string      `json:"brand,omitempty" validate:"omitempty,max=100"`
	Model               *string      `json:"model,omitempty" validate:"omitempty,max=100"`
	Description         *string      `json:"description,omitempty" validate:"omitempty,max=2000"`
	Price               *float64     `json:"price,omitempty" validate:"omitempty,min=0"`
	WarrantyMonths      *int         `json:"warranty_months,omitempty" validate:"omitempty,min=0,max=360"`
	WarrantyConditions  *string      `json:"warranty_conditions,omitempty" validate:"omitempty,max=1000"`
	ManufacturerWebsite *string      `json:"manufacturer_website,omitempty" validate:"omitempty,url"`
	Specifications      ProductSpecs `json:"specifications,omitempty"`
	Status              *string      `json:"status,omitempty" validate:"omitempty,oneof=active discontinued draft"`
}