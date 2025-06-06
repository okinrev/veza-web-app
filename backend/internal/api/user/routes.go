// internal/api/user/routes.go
package user

import (
    "net/http"
    "strconv"

    "github.com/okinrev/veza-web-app/internal/middleware"
    "github.com/okinrev/veza-web-app/internal/common"
    
    "github.com/gin-gonic/gin"
)

// SetupRoutes configure les routes utilisateur
func SetupRoutes(router *gin.RouterGroup, service *Service, jwtSecret string) {
    users := router.Group("/users")
    
    // Routes publiques
    users.GET("", getUsersHandler(service))
    users.GET("/:id", getUserByIDHandler(service))
    
    // Routes protégées
    protected := users.Group("")
    protected.Use(middleware.JWTAuthMiddleware(jwtSecret))
    {
        protected.GET("/profile", getProfileHandler(service))
        protected.PUT("/profile", updateProfileHandler(service))
        protected.POST("/change-password", changePasswordHandler(service))
        protected.GET("/stats", getUserStatsHandler(service))
    }
}

// getUsersHandler gère la récupération de la liste des utilisateurs
func getUsersHandler(service *Service) gin.HandlerFunc {
    return func(c *gin.Context) {
        page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
        limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
        search := c.Query("search")
        
        users, total, err := service.GetUsers(page, limit, search)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error":   "Failed to retrieve users",
                "success": false,
            })
            return
        }
        
        c.JSON(http.StatusOK, gin.H{
            "success": true,
            "data":    users,
            "meta": gin.H{
                "page":  page,
                "limit": limit,
                "total": total,
            },
        })
    }
}

// getUserByIDHandler gère la récupération d'un utilisateur par ID
func getUserByIDHandler(service *Service) gin.HandlerFunc {
    return func(c *gin.Context) {
        idStr := c.Param("id")
        id, err := strconv.Atoi(idStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{
                "error":   "Invalid user ID",
                "success": false,
            })
            return
        }
        
        user, err := service.GetUserByID(id)
        if err != nil {
            if err.Error() == "user not found" {
                c.JSON(http.StatusNotFound, gin.H{
                    "error":   "User not found",
                    "success": false,
                })
                return
            }
            c.JSON(http.StatusInternalServerError, gin.H{
                "error":   "Failed to retrieve user",
                "success": false,
            })
            return
        }
        
        c.JSON(http.StatusOK, gin.H{
            "success": true,
            "data":    user,
        })
    }
}

// getProfileHandler gère la récupération du profil de l'utilisateur connecté
func getProfileHandler(service *Service) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID, exists := common.GetUserIDFromContext(c)
        if !exists {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error":   "User ID not found in context",
                "success": false,
            })
            return
        }
        
        user, err := service.GetUserByID(userID)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error":   "Failed to retrieve profile",
                "success": false,
            })
            return
        }
        
        c.JSON(http.StatusOK, gin.H{
            "success": true,
            "data":    user,
        })
    }
}

// updateProfileHandler gère la mise à jour du profil
func updateProfileHandler(service *Service) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID, exists := common.GetUserIDFromContext(c)
        if !exists {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error":   "User ID not found in context",
                "success": false,
            })
            return
        }
        
        var req UpdateUserRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{
                "error":   "Invalid request data: " + err.Error(),
                "success": false,
            })
            return
        }
        
        user, err := service.UpdateUser(userID, req)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error":   "Failed to update profile: " + err.Error(),
                "success": false,
            })
            return
        }
        
        c.JSON(http.StatusOK, gin.H{
            "success": true,
            "data":    user,
            "message": "Profile updated successfully",
        })
    }
}

// changePasswordHandler gère le changement de mot de passe
func changePasswordHandler(service *Service) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID, exists := common.GetUserIDFromContext(c)
        if !exists {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error":   "User ID not found in context",
                "success": false,
            })
            return
        }
        
        var req struct {
            CurrentPassword string `json:"current_password" binding:"required"`
            NewPassword     string `json:"new_password" binding:"required,min=6"`
        }
        
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{
                "error":   "Invalid request data: " + err.Error(),
                "success": false,
            })
            return
        }
        
        err := service.ChangePassword(userID, req.CurrentPassword, req.NewPassword)
        if err != nil {
            if err.Error() == "current password is incorrect" {
                c.JSON(http.StatusBadRequest, gin.H{
                    "error":   "Current password is incorrect",
                    "success": false,
                })
                return
            }
            c.JSON(http.StatusInternalServerError, gin.H{
                "error":   "Failed to change password",
                "success": false,
            })
            return
        }
        
        c.JSON(http.StatusOK, gin.H{
            "success": true,
            "message": "Password changed successfully",
        })
    }
}

// getUserStatsHandler gère la récupération des statistiques utilisateur
func getUserStatsHandler(service *Service) gin.HandlerFunc {
    return func(c *gin.Context) {
        stats, err := service.GetUserStats()
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error":   "Failed to retrieve user statistics",
                "success": false,
            })
            return
        }
        
        c.JSON(http.StatusOK, gin.H{
            "success": true,
            "data":    stats,
        })
    }
}