// internal/routes/chat.go
package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/okinrev/veza-web-app/internal/handlers"
	"github.com/okinrev/veza-web-app/internal/middleware"
)

func (r *Router) setupChatRoutes(rg *gin.RouterGroup) {
	chatHandler := handlers.NewChatHandler(r.db)

	chat := rg.Group("/chat")
	chat.Use(middleware.JWTAuthMiddleware(r.jwtSecret)) // All chat routes require auth
	{
		// Direct messages
		dm := chat.Group("/dm")
		{
			dm.GET("/conversations", chatHandler.GetConversations)
			dm.GET("/unread-count", chatHandler.GetUnreadCount)
			dm.GET("/:user_id", chatHandler.GetDirectMessages)
			dm.POST("/:user_id", chatHandler.SendDirectMessage)
			dm.PUT("/:user_id/read", chatHandler.MarkAsRead)
		}

		// Message management
		messages := chat.Group("/messages")
		{
			messages.PUT("/:message_id", chatHandler.EditMessage)
			messages.DELETE("/:message_id", chatHandler.DeleteMessage)
		}
	}
}

func (r *Router) setupRoomRoutes(rg *gin.RouterGroup) {
	roomHandler := handlers.NewRoomHandler(r.db)

	rooms := rg.Group("/rooms")
	{
		// Public room discovery
		rooms.GET("/public", roomHandler.GetPublicRooms)
		rooms.GET("/:id", roomHandler.GetRoom)

		// Protected room operations
		protected := rooms.Group("/")
		protected.Use(middleware.JWTAuthMiddleware(r.jwtSecret))
		{
			// Room management
			protected.POST("", roomHandler.CreateRoom)
			protected.POST("/:id/join", roomHandler.JoinRoom)
			protected.POST("/:id/leave", roomHandler.LeaveRoom)

			// Room content
			protected.GET("/:id/members", roomHandler.GetRoomMembers)
			protected.GET("/:id/messages", roomHandler.GetRoomMessages)
			protected.POST("/:id/messages", roomHandler.SendRoomMessage)
		}
	}

	// Legacy chat routes for compatibility
	legacyChat := rg.Group("/chat")
	legacyChat.Use(middleware.JWTAuthMiddleware(r.jwtSecret))
	{
		legacyChat.GET("/rooms", roomHandler.GetPublicRooms)
		legacyChat.POST("/rooms", roomHandler.CreateRoom)
		legacyChat.GET("/rooms/:room/messages", roomHandler.GetRoomMessages)
		legacyChat.POST("/rooms/:room/messages", roomHandler.SendRoomMessage)
	}
}