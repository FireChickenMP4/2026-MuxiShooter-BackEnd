package controller

import (
	"MuXi/2026-MuxiShooter-Backend/dto"
	"MuXi/2026-MuxiShooter-Backend/middleware"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// @Summary		管理员按类型查询基础资源
// @Description	通过query参数type查询skills/achievements/items/cards；支持分页与可选id精确查询
// @Tags			admin-resource
// @Produce		json
// @Param			type		query		string												true	"资源类型(achievements/skills/items/cards)"
// @Param			id			query		int													false	"资源ID，传入后优先精确查询"
// @Param			name		query		string												false	"名称模糊搜索"
// @Param			skill_group	query		string												false	"技能组模糊搜索(type=skills有效)"
// @Param			page		query		int													false	"页码，默认1"
// @Param			page_size	query		int													false	"每页多少，默认20，最大100"
// @Success		200			{object}	dto.Response{data=dto.CommonAdminResourcePageData}	"查询成功"
// @Failure		400			{object}	dto.Response										"请求参数错误"
// @Failure		401			{object}	dto.Response										"登录状态异常"
// @Failure		500			{object}	dto.Response										"数据库查询失败"
// @Router			/api/admin/get/resources [get]
func GetResourcesByTypeForAdmin(c *gin.Context) {
	relationType, err := parseResourceType(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	handler, exists := adminResourceQueryHandlers[relationType]
	if !exists {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: ErrUnsupportedRelationType.Error()})
		return
	}

	pagination := middleware.GetPagination(c)
	list, total, err := handler(c, pagination)
	if err != nil {
		if errors.Is(err, ErrResourceIDInvalid) {
			c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.Response{Code: http.StatusInternalServerError, Message: "数据库查询失败：" + err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "查询成功",
		Data:    dto.CommonAdminResourcePageData{List: list, Total: total, Page: pagination.Page, PageSize: pagination.PageSize},
	})
}

// @Summary		管理员按类型创建基础资源
// @Description	通过query参数type创建skills/achievements/items/cards中的一种资源
// @Description	skills需要额外参数skill_group和prq_skill_id，其他资源只需要公共请求体
// @Tags			admin-resource
// @Accept			json
// @Produce		json
// @Param			type	query		string											true	"资源类型(achievements/skills/items/cards)"
// @Param			request	body		dto.CommonResourceCreateRequest					true	"创建请求体"
// @Success		200		{object}	dto.Response{data=dto.CommonAdminResourceData}	"创建成功"
// @Failure		400		{object}	dto.Response									"请求参数错误"
// @Failure		401		{object}	dto.Response									"登录状态异常"
// @Failure		409		{object}	dto.Response									"名称冲突"
// @Failure		500		{object}	dto.Response									"数据库错误"
// @Router			/api/admin/operation/resources [post]
func CreateResourceByTypeForAdmin(c *gin.Context) {
	relationType, err := parseResourceType(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	handler, exists := adminResourceCreateHandlers[relationType]
	if !exists {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: ErrUnsupportedRelationType.Error()})
		return
	}

	resource, err := handler(c)
	if err != nil {
		if errors.Is(err, ErrInvalidRequestBody) || errors.Is(err, ErrNoUpdateFields) {
			c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: err.Error()})
			return
		}
		if errors.Is(err, ErrResourceNameExists) {
			c.JSON(http.StatusConflict, dto.Response{Code: http.StatusConflict, Message: err.Error()})
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, dto.Response{Code: http.StatusNotFound, Message: "目标资源不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.Response{Code: http.StatusInternalServerError, Message: "创建失败：" + err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Code: http.StatusOK, Message: "创建成功", Data: resource})
}

// @Summary		管理员按类型更新基础资源
// @Description	通过query参数type更新skills/achievements/items/cards中的一种资源
// @Description	skills需要额外参数skill_group和prq_skill_id，其他资源只需要公共请求体
// @Tags			admin-resource
// @Accept			json
// @Produce		json
// @Param			type	query		string											true	"资源类型(achievements/skills/items/cards)"
// @Param			request	body		dto.CommonResourceUpdateRequest					true	"更新请求体"
// @Success		200		{object}	dto.Response{data=dto.CommonAdminResourceData}	"更新成功"
// @Failure		400		{object}	dto.Response									"请求参数错误"
// @Failure		401		{object}	dto.Response									"登录状态异常"
// @Failure		404		{object}	dto.Response									"目标资源不存在"
// @Failure		409		{object}	dto.Response									"名称冲突"
// @Failure		500		{object}	dto.Response									"数据库错误"
// @Router			/api/admin/update/resources [put]
func UpdateResourceByTypeForAdmin(c *gin.Context) {
	relationType, err := parseResourceType(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	handler, exists := adminResourceUpdateHandlers[relationType]
	if !exists {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: ErrUnsupportedRelationType.Error()})
		return
	}

	resource, err := handler(c)
	if err != nil {
		if errors.Is(err, ErrInvalidRequestBody) || errors.Is(err, ErrNoUpdateFields) {
			c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: err.Error()})
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, dto.Response{Code: http.StatusNotFound, Message: "目标资源不存在"})
			return
		}
		if errors.Is(err, ErrResourceNameExists) {
			c.JSON(http.StatusConflict, dto.Response{Code: http.StatusConflict, Message: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.Response{Code: http.StatusInternalServerError, Message: "更新失败：" + err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Code: http.StatusOK, Message: "更新成功", Data: resource})
}

// @Summary		管理员按类型删除基础资源
// @Description	通过query参数type删除skills/achievements/items/cards中的一种资源
// @Tags			admin-resource
// @Accept			json
// @Produce		json
// @Param			type	query		string									true	"资源类型(achievements/skills/items/cards)"
// @Param			request	body		dto.AdminDeleteResourceByTypeRequest	true	"删除请求体"
// @Success		200		{object}	dto.Response							"删除成功"
// @Failure		400		{object}	dto.Response							"请求参数错误"
// @Failure		401		{object}	dto.Response							"登录状态异常"
// @Failure		404		{object}	dto.Response							"目标资源不存在"
// @Failure		500		{object}	dto.Response							"数据库错误"
// @Router			/api/admin/operation/resources [delete]
func DeleteResourceByTypeForAdmin(c *gin.Context) {
	relationType, err := parseResourceType(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	var req dto.AdminDeleteResourceByTypeRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: "请求参数错误:" + err.Error()})
		return
	}

	handler, exists := adminResourceDeleteHandlers[relationType]
	if !exists {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: ErrUnsupportedRelationType.Error()})
		return
	}

	if err = handler(req.ID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, dto.Response{Code: http.StatusNotFound, Message: "目标资源不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.Response{Code: http.StatusInternalServerError, Message: "删除失败：" + err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Code: http.StatusOK, Message: "删除成功"})
}

// @Summary		管理员按类型查询任意用户关联数据
// @Description	通过query参数user_id和type查询指定用户在achievements/skills/items/cards中的关联数据
// @Description	skills会返回skill_grade，其他类型没有这个字段
// @Description	data.list: []dto.CommonUserRelationData
// @Tags			admin-resource
// @Produce		json
// @Param			user_id		query		int													true	"用户ID"
// @Param			type		query		string												true	"关联类型(achievements/skills/items/cards)"
// @Param			page		query		int													false	"页码，默认1"
// @Param			page_size	query		int													false	"每页多少，默认20，最大100"
// @Success		200			{object}	dto.Response{data=dto.CommonUserRelationPageData}	"查询成功"
// @Failure		400			{object}	dto.Response										"请求参数错误"
// @Failure		401			{object}	dto.Response										"登录状态异常"
// @Failure		500			{object}	dto.Response										"数据库查询失败"
// @Router			/api/admin/get/user-relations [get]
func GetUserRelationsByTypeForAdmin(c *gin.Context) {
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: "缺少user_id参数"})
		return
	}

	parsedID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil || parsedID == 0 {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: "user_id参数格式错误"})
		return
	}

	relationTypeStr := c.Query("type")
	if relationTypeStr == "" {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: "缺少type参数"})
		return
	}

	relationType, err := ParseUserRelationType(relationTypeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	pagination := middleware.GetPagination(c)
	list, total, err := QueryUserRelationByTypeWithUserID(uint(parsedID), relationType, pagination)
	if err != nil {
		if errors.Is(err, ErrUnsupportedRelationType) || errors.Is(err, ErrUserIDTypeInvalid) {
			c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.Response{Code: http.StatusInternalServerError, Message: "数据库查询失败：" + err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "查询成功",
		Data:    dto.CommonUserRelationPageData{List: list, Total: total, Page: pagination.Page, PageSize: pagination.PageSize},
	})
}
