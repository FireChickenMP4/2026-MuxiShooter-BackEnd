package controller

import (
	config "MuXi/2026-MuxiShooter-Backend/config"
	"MuXi/2026-MuxiShooter-Backend/dto"
	"MuXi/2026-MuxiShooter-Backend/middleware"
	models "MuXi/2026-MuxiShooter-Backend/models"
	utils "MuXi/2026-MuxiShooter-Backend/utils"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const initialAdminUserID uint = 1

// @Summary		获取用户列表
// @Description	注: 管理员可用，查询结果是多重模糊搜索叠加的效果
// @Description	以及页码不输入或不合规范自动为第一页，每页多少不输入默认20，最多100
// @Description	如果查询结果不存在则返回切片为空
// @Description	用了id查询的话就一定只是一个确定的，而不是模糊搜索，其他参数就没用了（分页也是）
// @Tags			admin-user
// @Produce		json
// @Param			user_id		query		int									false	"用户id"
// @Param			username	query		string								false	"用户名"
// @Param			group		query		string								false	"权限组(user/admin)"
// @Param			page		query		int									false	"页码，默认1"
// @Param			page_size	query		int									false	"每页多少，默认20，最大100"
// @Success		200			{object}	dto.Response{data=dto.PaginatedData}	"查询成功"
// @Failure		401			{object}	dto.Response							"登录状态异常"
// @Failure		500			{object}	dto.Response							"数据库查询失败"
// @Router			/api/admin/get/getusers [get]
func GetUsers(c *gin.Context) {
	var err error
	pagination := middleware.GetPagination(c)

	id := c.Query("user_id")
	username := utils.SqlSafeLikeKeyword(c.Query("username"))
	group := utils.SqlSafeLikeKeyword(c.Query("group"))

	var users []models.User
	var total int64

	if id != "" {
		var user models.User
		pagination = models.Pagination{Page: config.DefaultPage, PageSize: config.DefaultPageSize, Limit: config.DefaultPageSize, Offset: 0}
		result := currentDB().First(&user, id)
		err = result.Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				total = 0
			} else {
				c.JSON(http.StatusInternalServerError, dto.Response{Code: http.StatusInternalServerError, Message: "数据库查询失败：" + err.Error()})
				return
			}
		} else {
			total = 1
		}
		if total != 0 {
			users = append(users, user)
		}
	} else {
		result := currentDB().Model(&models.User{})
		if username != "" {
			result = result.Where("username LIKE ?", "%"+username+"%")
		}
		if group != "" {
			result = result.Where("group LIKE ?", "%"+group+"%")
		}
		result.Count(&total)
		result.Limit(pagination.Limit).Offset(pagination.Offset).Find(&users)
	}

	c.JSON(http.StatusOK, dto.Response{Code: http.StatusOK, Message: "查询成功", Data: dto.PaginatedData{List: users, Total: total, Page: pagination.Page, PageSize: pagination.PageSize}})
}

// @Summary		管理员删除用户
// @Description	管理员删除用户。ID=1(初始化管理员)不可删除；其他管理员仅可删除普通用户
// @Tags			admin-user
// @Accept			json
// @Produce		json
// @Param			request	body		dto.AdminDeleteUserRequest	true	"删除用户请求"
// @Success		200		{object}	dto.Response				"删除成功"
// @Failure		400		{object}	dto.Response				"请求参数错误"
// @Failure		401		{object}	dto.Response				"登录状态异常"
// @Failure		403		{object}	dto.Response				"权限不足"
// @Failure		404		{object}	dto.Response				"用户不存在"
// @Failure		500		{object}	dto.Response				"数据库错误"
// @Router			/api/admin/operation/deleteuser [delete]
func DeleteUserByAdmin(c *gin.Context) {
	var req dto.AdminDeleteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: "请求参数错误:" + err.Error()})
		return
	}

	operatorID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.Response{Code: http.StatusUnauthorized, Message: "解析后token中缺少用户信息"})
		return
	}
	adminID, ok := operatorID.(uint)
	if !ok || adminID == 0 {
		c.JSON(http.StatusUnauthorized, dto.Response{Code: http.StatusUnauthorized, Message: "用户信息格式错误"})
		return
	}

	if req.UserID == initialAdminUserID {
		c.JSON(http.StatusForbidden, dto.Response{Code: http.StatusForbidden, Message: "初始化管理员不可删除"})
		return
	}

	var targetUser models.User
	if err := currentDB().First(&targetUser, req.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, dto.Response{Code: http.StatusNotFound, Message: "目标用户不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.Response{Code: http.StatusInternalServerError, Message: "数据库查询失败：" + err.Error()})
		return
	}

	if adminID != initialAdminUserID && targetUser.Group != "user" {
		c.JSON(http.StatusForbidden, dto.Response{Code: http.StatusForbidden, Message: "仅初始化管理员可删除管理员账户"})
		return
	}

	tx := currentDB().Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{Code: http.StatusInternalServerError, Message: "开启事务失败：" + tx.Error.Error()})
		return
	}

	if err := tx.Delete(&targetUser).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, dto.Response{Code: http.StatusInternalServerError, Message: "删除用户失败：" + err.Error()})
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, dto.Response{Code: http.StatusInternalServerError, Message: "提交删除请求失败：" + err.Error()})
		return
	}

	if targetUser.HeadImagePath != "" && targetUser.HeadImagePath != config.DefaultHeadImagePath {
		if removeErr := utils.RemoveFile(targetUser.HeadImagePath); removeErr != nil {
			log.Printf("删除用户头像文件失败(user_id:%d,path:%s): %v", targetUser.ID, targetUser.HeadImagePath, removeErr)
		}
	}

	c.JSON(http.StatusOK, dto.Response{Code: http.StatusOK, Message: "删除用户成功"})
}

// @Summary		管理员修改用户权限组
// @Description	ID=1(初始化管理员)权限组不可修改；仅ID=1可修改其他用户权限组
// @Tags			admin-user
// @Accept			json
// @Produce		json
// @Param			request	body		dto.AdminUpdateUserGroupRequest	true	"修改权限组请求"
// @Success		200		{object}	dto.Response					"修改成功"
// @Failure		400		{object}	dto.Response					"请求参数错误"
// @Failure		401		{object}	dto.Response					"登录状态异常"
// @Failure		403		{object}	dto.Response					"权限不足"
// @Failure		404		{object}	dto.Response					"用户不存在"
// @Failure		500		{object}	dto.Response					"数据库错误"
// @Router			/api/admin/update/usergroup [put]
func UpdateUserGroupByAdmin(c *gin.Context) {
	var req dto.AdminUpdateUserGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{Code: http.StatusBadRequest, Message: "请求参数错误:" + err.Error()})
		return
	}

	operatorID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.Response{Code: http.StatusUnauthorized, Message: "解析后token中缺少用户信息"})
		return
	}
	adminID, ok := operatorID.(uint)
	if !ok || adminID == 0 {
		c.JSON(http.StatusUnauthorized, dto.Response{Code: http.StatusUnauthorized, Message: "用户信息格式错误"})
		return
	}

	if req.UserID == initialAdminUserID {
		c.JSON(http.StatusForbidden, dto.Response{Code: http.StatusForbidden, Message: "初始化管理员权限组不可修改"})
		return
	}

	if adminID != initialAdminUserID {
		c.JSON(http.StatusForbidden, dto.Response{Code: http.StatusForbidden, Message: "仅初始化管理员可修改用户权限组"})
		return
	}

	var targetUser models.User
	tx := currentDB().Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{Code: http.StatusInternalServerError, Message: "开启事务失败：" + tx.Error.Error()})
		return
	}

	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&targetUser, req.UserID).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, dto.Response{Code: http.StatusNotFound, Message: "目标用户不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.Response{Code: http.StatusInternalServerError, Message: "数据库查询失败：" + err.Error()})
		return
	}

	if err := tx.Model(&targetUser).Update("group", req.NewGroup).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, dto.Response{Code: http.StatusInternalServerError, Message: "修改权限组失败：" + err.Error()})
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, dto.Response{Code: http.StatusInternalServerError, Message: "提交修改请求失败：" + err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Code: http.StatusOK, Message: "修改用户权限组成功"})
}
