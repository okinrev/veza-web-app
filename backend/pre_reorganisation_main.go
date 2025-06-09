//file: backend/main.go

package main

import (
	"log"
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/okinrev/veza-web-app/internal/api/admin/routes"
	"github.com/okinrev/veza-web-app/internal/api/auth/routes"
	"github.com/okinrev/veza-web-app/internal/api/file/routes"
	"github.com/okinrev/veza-web-app/internal/api/message/routes"
	"github.com/okinrev/veza-web-app/internal/api/ressource/routes"
	"github.com/okinrev/veza-web-app/internal/api/room/routes"
	"github.com/okinrev/veza-web-app/internal/api/track/routes"
	"github.com/okinrev/veza-web-app/internal/api/user/routes"
)

func main() {
	db.Init()
	r := mux.NewRouter()

	r.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("../frontend", "favicon.ico"))
	})

	routes.RegisterAuthRoutes(r)

	routes.RegisterAdminRoutes(r)

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
