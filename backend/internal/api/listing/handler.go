//file: backend/handlers/listing.go

package handlers

import (
    "database/sql"
    "encoding/json"
    "net/http"
	"log"

	"github.com/lib/pq"
    "github.com/gorilla/mux"
    "backend/models"
    "backend/db"
)

func CreateListing(w http.ResponseWriter, r *http.Request) {
    var listing models.Listing
    if err := json.NewDecoder(r.Body).Decode(&listing); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

	listing.Status = "open"

	query := `INSERT INTO listings (user_id, product_id, description, state, price, exchange_for, images, status)
	VALUES ($1, $2, $3, $4, $5, $6, $7, 'open') RETURNING id, status, created_at`

	err := db.DB.QueryRow(query, listing.UserID, listing.ProductID, listing.Description,
	listing.State, listing.Price, listing.ExchangeFor, pq.Array(listing.Images)).
	Scan(&listing.ID, &listing.Status, &listing.CreatedAt)
    if err != nil {
		log.Printf("❌ Erreur JSON côté backend: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(listing)
}

func GetAllListings(w http.ResponseWriter, r *http.Request) {
    rows, err := db.DB.Query(`SELECT id, user_id, product_id, description, state, price, exchange_for, images, status, created_at FROM listings WHERE status = 'open'`)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var listings []models.Listing
    for rows.Next() {
        var l models.Listing
        if err := rows.Scan(&l.ID, &l.UserID, &l.ProductID, &l.Description, &l.State, &l.Price, &l.ExchangeFor, pq.Array(&l.Images), &l.Status, &l.CreatedAt); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        listings = append(listings, l)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(listings)
}

func GetListingByID(w http.ResponseWriter, r *http.Request) {
    id := mux.Vars(r)["id"]
    var l models.Listing
    query := `SELECT id, user_id, product_id, description, state, price, exchange_for, images, status, created_at FROM listings WHERE id = $1`
    err := db.DB.QueryRow(query, id).Scan(&l.ID, &l.UserID, &l.ProductID, &l.Description, &l.State, &l.Price, &l.ExchangeFor, pq.Array(&l.Images), &l.Status, &l.CreatedAt)
    if err == sql.ErrNoRows {
        http.Error(w, "Annonce non trouvée", http.StatusNotFound)
        return
    } else if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(l)
}

func DeleteListing(w http.ResponseWriter, r *http.Request) {
    id := mux.Vars(r)["id"]
    _, err := db.DB.Exec(`DELETE FROM listings WHERE id = $1`, id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusNoContent)
}
