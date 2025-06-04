//file: backend/routes/track.go

package routes

import (
	"github.com/gorilla/mux"
	"backend/handlers"
	"backend/middleware"
	"backend/utils"
	"net/http"
)

func RegisterTrackRoutes(r *mux.Router) {

	r.Handle("/tracks", middleware.JWTAuthMiddleware(http.HandlerFunc(handlers.AddTrackWithUpload))).Methods("POST")
	r.Handle("/tracks", middleware.JWTAuthMiddleware(http.HandlerFunc(handlers.ListTracks))).Methods("GET")
	r.Handle("/generate-stream-url", middleware.JWTAuthMiddleware(http.HandlerFunc(utils.HandleGenerateSignedURL))).Methods("GET")
	r.HandleFunc("/stream/{filename}", utils.StreamAudioWithValidation).Methods("GET")


}