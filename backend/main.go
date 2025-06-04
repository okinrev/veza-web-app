//file: backend/main.go

package main

import (
	"log"
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
	"veza-backend/db"
	"veza-backend/routes"
)

func main() {
	db.Init()
	r := mux.NewRouter()

	r.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("../frontend", "favicon.ico"))
	})
	

	routes.RegisterAuthRoutes(r)

	routes.RegisterProductRoutes(r)

	routes.RegisterUserProductRoutes(r)

	routes.RegisterFileRoutes(r)

	routes.RegisterRessourceRoutes(r)

	routes.RegisterMessageRoutes(r, db.DB)

	routes.RegisterRoomRoutes(r, db.DB)

	routes.RegisterUserRoutes(r)

	routes.RegisterTrackRoutes(r)	

	routes.RegisterSharedRoutes(r)

	routes.RegisterTagRoutes(r)

	routes.RegisterSuggestionRoutes(r)

	routes.RegisterSearchRoutes(r, db.DB)

	routes.RegisterListingRoutes(r)

	routes.RegisterOfferRoutes(r)

	routes.RegisterAdminRoutes(r)


	r.PathPrefix("/").Handler(http.FileServer(http.Dir("../frontend")))

	log.Println("Serveur lancé sur :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
