//file: backend/handlers/offer.go

package handlers

import (
    "encoding/json"
    "net/http"

    "github.com/gorilla/mux"
    "veza-web-app/models"
    "veza-web-app/db"
)

func CreateOffer(w http.ResponseWriter, r *http.Request) {
    var offer models.Offer
    if err := json.NewDecoder(r.Body).Decode(&offer); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    query := `INSERT INTO offers (listing_id, from_user_id, proposed_product_id, message, status) 
              VALUES ($1, $2, $3, $4, 'pending') RETURNING id`
    err := db.DB.QueryRow(query, offer.ListingID, offer.FromUserID, offer.ProposedProductID, offer.Message).Scan(&offer.ID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(offer)
}

func AcceptOffer(w http.ResponseWriter, r *http.Request) {
    id := mux.Vars(r)["id"]

    tx, err := db.DB.Begin()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer tx.Rollback()

    // Mettre à jour l'offre comme acceptée
    _, err = tx.Exec(`UPDATE offers SET status = 'accepted' WHERE id = $1`, id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Fermer l'annonce liée
    _, err = tx.Exec(`UPDATE listings SET status = 'closed' WHERE id = (SELECT listing_id FROM offers WHERE id = $1)`, id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if err = tx.Commit(); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}
