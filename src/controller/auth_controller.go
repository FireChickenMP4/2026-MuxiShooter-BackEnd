package controller

import (
	config "MuXi/2026-MuxiShooter-Backend/config"
	"MuXi/2026-MuxiShooter-Backend/dto"
	models "MuXi/2026-MuxiShooter-Backend/models"
	utils "MuXi/2026-MuxiShooter-Backend/utils"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// @Summary		用户注册
// @Description	注册用户
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			request	body		dto.RegisterRequest				true	"注册请求"
// @Success		200		{object}	dto.Response{data=dto.AuthData}	"注册成功"
// @Failure		400		{object}	dto.Response					"请求参数错误"
// @Failure		409		{object}	dto.Response					"用户已存在"
// @Failure		500		{object}	dto.Response					"服务器错误"
// @Router			/api/auth/register [post]
func Register(c *gin.Context) {
	var req dto.RegisterRequest
	var err error

	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误:" + err.Error(),
		})
		return
	}

	var searchedUser models.User
	err = currentDB().Where("username = ?", req.UserName).First(&searchedUser).Error
	if err == nil {
		c.JSON(http.StatusConflict, dto.Response{
			Code:    http.StatusConflict,
			Message: "用户已存在",
		})
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "查询数据库失败：" + err.Error(),
		})
		return
	}

	hashedPsw, err := utils.Hashtool(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "注册密码哈希失败：" + err.Error(),
		})
		return
	}

	newUser := models.User{
		Username:      req.UserName,
		Password:      hashedPsw,
		Group:         "user",
		HeadImagePath: config.DefaultHeadImagePath,
	}

	if err = currentDB().Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "注册用户失败：" + err.Error(),
		})
		return
	}

	token, expirationTime, err := utils.GenerateToken(newUser, currentJWTSecret())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "注册用户成功",
		Data: dto.AuthData{
			User: dto.CommonUserData{
				UserID:        newUser.ID,
				Username:      newUser.Username,
				Group:         newUser.Group,
				HeadImagePath: newUser.HeadImagePath,
				StrengthCoin:  newUser.StrengthCoin,
				SelectCoin:    newUser.SelectCoin,
			},
			Token:     token,
			ExpiresAt: expirationTime.Unix(),
		},
	})
}

// @Summary		用户登录
// @Description	用户登录
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			request	body		dto.LoginRequest				true	"注册请求"
// @Success		200		{object}	dto.Response{data=dto.AuthData}	"登录成功"
// @Failure		400		{object}	dto.Response					"请求参数错误"
// @Failure		403		{object}	dto.Response					"认证失败"
// @Failure		500		{object}	dto.Response					"服务器错误"
// @Router			/api/auth/login [post]
func Login(c *gin.Context) {
	var req dto.LoginRequest
	var err error

	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误:" + err.Error(),
		})
		return
	}

	var user models.User
	err = currentDB().Where("username = ?", req.UserName).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusForbidden, dto.Response{
			Code:    http.StatusForbidden,
			Message: "用户不存在",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "查询数据库失败：" + err.Error(),
		})
		return
	}

	err = utils.ComparePassword(user.Password, req.Password)
	if err != nil {
		c.JSON(http.StatusForbidden, dto.Response{
			Code:    http.StatusForbidden,
			Message: "密码错误",
		})
		return
	}

	token, expirationTime, err := utils.GenerateToken(user, currentJWTSecret())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "登录成功",
		Data: dto.AuthData{
			User: dto.CommonUserData{
				UserID:        user.ID,
				Username:      user.Username,
				Group:         user.Group,
				HeadImagePath: user.HeadImagePath,
				StrengthCoin:  user.StrengthCoin,
				SelectCoin:    user.SelectCoin,
			},
			Token:     token,
			ExpiresAt: expirationTime.Unix(),
		},
	})
}
