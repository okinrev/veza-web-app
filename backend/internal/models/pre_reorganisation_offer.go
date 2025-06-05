//file: backend/models/offer.go

package models

type Offer struct {
    ID                int     `db:"id" json:"id"`
    ListingID         int     `db:"listing_id" json:"listing_id"`
    FromUserID        int     `db:"from_user_id" json:"from_user_id"`
    ProposedProductID int     `db:"proposed_product_id" json:"proposed_product_id"`
    Message           *string `db:"message" json:"message,omitempty"`
    Status            string  `db:"status" json:"status"`
    CreatedAt         string  `db:"created_at" json:"created_at"`
}