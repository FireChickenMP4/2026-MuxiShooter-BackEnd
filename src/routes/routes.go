package routes

import (
	"MuXi/2026-MuxiShooter-Backend/controller"
	"MuXi/2026-MuxiShooter-Backend/dto"
	"MuXi/2026-MuxiShooter-Backend/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHTTPHandler interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
}

type ProfileHTTPHandler interface {
	Logout(c *gin.Context)
	UpdatePassword(c *gin.Context)
	UpdateUsername(c *gin.Context)
	UpdateHeadImage(c *gin.Context)
	UpdateCoinByType(c *gin.Context)
	CreateSelfRelationByType(c *gin.Context)
	UpdateSelfRelationByType(c *gin.Context)
	DeleteSelfRelationByType(c *gin.Context)
	GetSelfProfile(c *gin.Context)
	GetSelfRelationsByType(c *gin.Context)
}

func RegisterRoutes(r *gin.Engine, authHandler AuthHTTPHandler, profileHandler ProfileHTTPHandler, jwtAuthMiddleware gin.HandlerFunc) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, dto.Response{
			Code:    http.StatusOK, //200
			Message: "I'm OK.",
		})
	})
	if authHandler == nil {
		panic("auth handler is nil")
	}
	if profileHandler == nil {
		panic("profile handler is nil")
	}
	if jwtAuthMiddleware == nil {
		panic("jwt auth middleware is nil")
	}
	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		authGroup := api.Group("/")
		authGroup.Use(jwtAuthMiddleware)
		{
			profile := authGroup.Group("/profile")
			{
				update := profile.Group("/update")
				{
					update.PUT("/password", profileHandler.UpdatePassword)
					update.PUT("/username", profileHandler.UpdateUsername)
					update.PUT("/headimage", profileHandler.UpdateHeadImage)
					update.PUT("/coin", profileHandler.UpdateCoinByType)
					update.PUT("/relations", profileHandler.UpdateSelfRelationByType)
				}
				operation := profile.Group("/operation")
				{
					operation.GET("/logout", profileHandler.Logout)
					operation.POST("/relations", profileHandler.CreateSelfRelationByType)
					operation.DELETE("/relations", profileHandler.DeleteSelfRelationByType)
				}
				get := profile.Group("/get")
				{
					get.GET("/self", profileHandler.GetSelfProfile)

					paginatedGet := get.Group("/")
					paginatedGet.Use(middleware.PaginationMiddleware())
					{
						paginatedGet.GET("/relations", profileHandler.GetSelfRelationsByType)
					}
				}
			}

			adminGroup := authGroup.Group("/admin")
			adminGroup.Use(middleware.AdminRequired())
			{
				operationGroup := adminGroup.Group("/operation")
				{
					operationGroup.DELETE("/deleteuser", controller.DeleteUserByAdmin)
					operationGroup.POST("/resources", controller.CreateResourceByTypeForAdmin)
					operationGroup.DELETE("/resources", controller.DeleteResourceByTypeForAdmin)
				}

				updateGroup := adminGroup.Group("/update")
				{
					updateGroup.PUT("/usergroup", controller.UpdateUserGroupByAdmin)
					updateGroup.PUT("/resources", controller.UpdateResourceByTypeForAdmin)
				}

				getGroup := adminGroup.Group("/get")
				{
					paginatedGroup := getGroup.Group("/")
					paginatedGroup.Use(middleware.PaginationMiddleware())
					{
						paginatedGroup.GET("/getusers", controller.GetUsers)
						paginatedGroup.GET("/resources", controller.GetResourcesByTypeForAdmin)
						paginatedGroup.GET("/user-relations", controller.GetUserRelationsByTypeForAdmin)
					}
				}
			}
		}
	}
}
