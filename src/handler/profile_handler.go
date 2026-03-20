package handler

import (
	config "MuXi/2026-MuxiShooter-Backend/config"
	"MuXi/2026-MuxiShooter-Backend/dto"
	"MuXi/2026-MuxiShooter-Backend/middleware"
	"MuXi/2026-MuxiShooter-Backend/service"
	"MuXi/2026-MuxiShooter-Backend/utils"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProfileHandler struct {
	profileService *service.ProfileService
}

func NewProfileHandler(profileService *service.ProfileService) *ProfileHandler {
	return &ProfileHandler{profileService: profileService}
}

func (h *ProfileHandler) Logout(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.Response{
			Code:    http.StatusUnauthorized,
			Message: service.ErrMissingUserContext.Error(),
		})
		return
	}

	err := h.profileService.Logout(userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) || errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, dto.Response{
				Code:    http.StatusUnauthorized,
				Message: "用户不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("服务器错误: %v, 请重试", err),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "登出成功",
	})
}

func (h *ProfileHandler) UpdatePassword(c *gin.Context) {
	var req dto.UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误:" + err.Error(),
		})
		return
	}

	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.Response{
			Code:    http.StatusUnauthorized,
			Message: service.ErrMissingUserContext.Error(),
		})
		return
	}

	err := h.profileService.UpdatePassword(userID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSamePassword):
			c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: err.Error()})
		case errors.Is(err, service.ErrUserNotFound), errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, dto.Response{Code: http.StatusNotFound, Message: "用户不存在"})
		case errors.Is(err, service.ErrPasswordTooFrequent):
			c.JSON(http.StatusForbidden, dto.Response{Code: http.StatusForbidden, Message: err.Error()})
		case errors.Is(err, service.ErrInvalidOldPassword):
			c.JSON(http.StatusForbidden, dto.Response{Code: http.StatusForbidden, Message: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, dto.Response{Code: http.StatusInternalServerError, Message: "服务器错误:" + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "修改密码成功，已退出登录，请重新登陆",
	})
}

func (h *ProfileHandler) UpdateUsername(c *gin.Context) {
	var req dto.UpdateUsernameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误:" + err.Error(),
		})
		return
	}

	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.Response{
			Code:    http.StatusUnauthorized,
			Message: service.ErrMissingUserContext.Error(),
		})
		return
	}

	err := h.profileService.UpdateUsername(userID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound), errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, dto.Response{Code: http.StatusNotFound, Message: "用户不存在"})
		case errors.Is(err, service.ErrUsernameTooFrequent):
			c.JSON(http.StatusForbidden, dto.Response{Code: http.StatusForbidden, Message: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, dto.Response{Code: http.StatusInternalServerError, Message: "服务器错误:" + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "修改用户名成功",
	})
}

func (h *ProfileHandler) UpdateHeadImage(c *gin.Context) {
	var req dto.UpdateHeadImageRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误:" + err.Error(),
		})
		return
	}

	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.Response{
			Code:    http.StatusUnauthorized,
			Message: service.ErrMissingUserContext.Error(),
		})
		return
	}

	if req.NewHeadImage == nil || req.NewHeadImage.Size == 0 {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "头像为空",
		})
		return
	}

	log.Printf("用户(id:%d)上传头像,Size:%d", userID, req.NewHeadImage.Size)
	savePath, err := utils.SaveImages(c, req.NewHeadImage, config.PrefixHeadImg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "图片保存失败：" + err.Error(),
		})
		return
	}

	oldHeadImagePath, err := h.profileService.UpdateHeadImage(userID, savePath)
	if err != nil {
		_ = utils.RemoveFile(savePath)
		switch {
		case errors.Is(err, service.ErrUserNotFound), errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, dto.Response{Code: http.StatusNotFound, Message: "用户不存在"})
		case errors.Is(err, service.ErrHeadImageTooFrequent):
			c.JSON(http.StatusForbidden, dto.Response{Code: http.StatusForbidden, Message: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, dto.Response{Code: http.StatusInternalServerError, Message: "修改头像失败：" + err.Error()})
		}
		return
	}

	if oldHeadImagePath != "" && oldHeadImagePath != savePath && oldHeadImagePath != config.DefaultHeadImagePath {
		if removeErr := utils.RemoveFile(oldHeadImagePath); removeErr != nil {
			log.Printf("删除旧头像失败(user_id:%d,path:%s): %v", userID, oldHeadImagePath, removeErr)
		}
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "修改头像成功",
	})
}

func (h *ProfileHandler) UpdateCoinByType(c *gin.Context) {
	var req dto.UpdateCoinByTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误:" + err.Error(),
		})
		return
	}
	if req.Coin == nil {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: "coin不能为空"})
		return
	}

	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.Response{
			Code:    http.StatusUnauthorized,
			Message: service.ErrMissingUserContext.Error(),
		})
		return
	}

	coinType := c.Query("type")
	userData, err := h.profileService.UpdateCoinByType(userID, coinType, *req.Coin)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMissingCoinType), errors.Is(err, service.ErrUnsupportedCoinType):
			c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: err.Error()})
		case errors.Is(err, service.ErrUserNotFound), errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, dto.Response{Code: http.StatusNotFound, Message: "用户不存在"})
		default:
			c.JSON(http.StatusInternalServerError, dto.Response{Code: http.StatusInternalServerError, Message: "数据库查询失败：" + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "修改金币成功",
		Data:    userData,
	})
}

func (h *ProfileHandler) CreateSelfRelationByType(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.Response{Code: http.StatusUnauthorized, Message: service.ErrMissingUserContext.Error()})
		return
	}

	relationTypeStr := c.Query("type")
	if relationTypeStr == "" {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: "缺少type参数"})
		return
	}
	relationType, err := service.ParseUserRelationType(relationTypeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	var req dto.UserRelationCreateRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: "请求参数错误:" + err.Error()})
		return
	}

	data, err := h.profileService.CreateSelfRelationByType(userID, relationType, req.ResourceID)
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, dto.Response{Code: http.StatusNotFound, Message: "目标资源不存在"})
		case errors.Is(err, service.ErrResourceNameExists):
			c.JSON(http.StatusConflict, dto.Response{Code: http.StatusConflict, Message: "关联已存在"})
		default:
			c.JSON(http.StatusInternalServerError, dto.Response{Code: http.StatusInternalServerError, Message: "创建失败：" + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, dto.Response{Code: http.StatusOK, Message: "创建成功", Data: data})
}

func (h *ProfileHandler) UpdateSelfRelationByType(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.Response{Code: http.StatusUnauthorized, Message: service.ErrMissingUserContext.Error()})
		return
	}

	relationTypeStr := c.Query("type")
	if relationTypeStr == "" {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: "缺少type参数"})
		return
	}
	relationType, err := service.ParseUserRelationType(relationTypeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	var req dto.UserRelationUpdateRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: "请求参数错误:" + err.Error()})
		return
	}

	data, err := h.profileService.UpdateSelfRelationByType(userID, relationType, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNoUpdateFields), errors.Is(err, service.ErrSkillGradeOnlyForSkills), errors.Is(err, service.ErrUnsupportedRelationType):
			c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: err.Error()})
		case errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, dto.Response{Code: http.StatusNotFound, Message: "关联记录不存在"})
		default:
			c.JSON(http.StatusInternalServerError, dto.Response{Code: http.StatusInternalServerError, Message: "更新失败：" + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, dto.Response{Code: http.StatusOK, Message: "更新成功", Data: data})
}

func (h *ProfileHandler) DeleteSelfRelationByType(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.Response{Code: http.StatusUnauthorized, Message: service.ErrMissingUserContext.Error()})
		return
	}

	relationTypeStr := c.Query("type")
	if relationTypeStr == "" {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: "缺少type参数"})
		return
	}
	relationType, err := service.ParseUserRelationType(relationTypeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	var req dto.UserRelationDeleteRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: "请求参数错误:" + err.Error()})
		return
	}

	if err = h.profileService.DeleteSelfRelationByType(userID, relationType, req.ResourceID); err != nil {
		switch {
		case errors.Is(err, service.ErrUnsupportedRelationType):
			c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: err.Error()})
		case errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, dto.Response{Code: http.StatusNotFound, Message: "关联记录不存在"})
		default:
			c.JSON(http.StatusInternalServerError, dto.Response{Code: http.StatusInternalServerError, Message: "删除失败：" + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, dto.Response{Code: http.StatusOK, Message: "删除成功"})
}

func (h *ProfileHandler) GetSelfProfile(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.Response{Code: http.StatusUnauthorized, Message: service.ErrMissingUserContext.Error()})
		return
	}

	data, err := h.profileService.GetSelfProfile(userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound), errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, dto.Response{Code: http.StatusNotFound, Message: "用户不存在"})
		default:
			c.JSON(http.StatusInternalServerError, dto.Response{Code: http.StatusInternalServerError, Message: "数据库查询失败：" + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, dto.Response{Code: http.StatusOK, Message: "查询成功", Data: data})
}

func (h *ProfileHandler) GetSelfRelationsByType(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.Response{Code: http.StatusUnauthorized, Message: service.ErrMissingUserContext.Error()})
		return
	}

	pagination := middleware.GetPagination(c)
	list, total, err := h.profileService.GetSelfRelationsByType(userID, c.Query("type"), pagination)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMissingRelationType), errors.Is(err, service.ErrUnsupportedRelationType):
			c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, dto.Response{Code: http.StatusInternalServerError, Message: "数据库查询失败：" + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "查询成功",
		Data: dto.CommonUserRelationPageData{
			List:     list,
			Total:    total,
			Page:     pagination.Page,
			PageSize: pagination.PageSize,
		},
	})
}

func getUserIDFromContext(c *gin.Context) (uint, bool) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	userID, ok := userIDValue.(uint)
	if !ok {
		return 0, false
	}
	return userID, true
}
