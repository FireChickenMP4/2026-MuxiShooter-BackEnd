package controller

import (
	config "MuXi/2026-MuxiShooter-Backend/config"
	"MuXi/2026-MuxiShooter-Backend/dto"
	"MuXi/2026-MuxiShooter-Backend/middleware"
	models "MuXi/2026-MuxiShooter-Backend/models"
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var (
	ErrUnsupportedRelationType = errors.New("不支持的关联表类型")
	ErrUserIDMissing           = errors.New("上下文中缺少用户ID")
	ErrUserIDTypeInvalid       = errors.New("上下文中的用户ID格式错误")
)

type UserRelationType string
type relationQueryHandler func(userID uint, pagination models.Pagination) ([]dto.CommonUserRelationData, int64, error)

const (
	UserRelationAchievement UserRelationType = "achievements"
	UserRelationSkill       UserRelationType = "skills"
	UserRelationItem        UserRelationType = "items"
	UserRelationCard        UserRelationType = "cards"
)

var relationQueryHandlers = map[UserRelationType]relationQueryHandler{
	UserRelationAchievement: queryUserAchievements,
	UserRelationSkill:       queryUserSkills,
	UserRelationItem:        queryUserItems,
	UserRelationCard:        queryUserCards,
}

func ParseUserRelationType(val string) (UserRelationType, error) {
	relationType := UserRelationType(val)
	if _, exists := relationQueryHandlers[relationType]; exists {
		return relationType, nil
	}
	return "", ErrUnsupportedRelationType
}

// QueryUserRelationByType 查询用户在指定关联表中的记录。
// 返回值依次为：查询结果切片、总条数、错误。
func QueryUserRelationByType(c *gin.Context, relationType UserRelationType) ([]dto.CommonUserRelationData, int64, error) {
	userID, err := getRelationQueryUserID(c)
	if err != nil {
		return nil, 0, err
	}
	pagination := middleware.GetPagination(c)
	return QueryUserRelationByTypeWithUserID(userID, relationType, pagination)
}

func QueryUserRelationByTypeWithUserID(userID uint, relationType UserRelationType, pagination models.Pagination) ([]dto.CommonUserRelationData, int64, error) {
	if userID == 0 {
		return nil, 0, ErrUserIDTypeInvalid
	}

	handler, exists := relationQueryHandlers[relationType]
	if !exists {
		return nil, 0, ErrUnsupportedRelationType
	}

	return handler(userID, pagination)
}

func queryUserAchievements(userID uint, pagination models.Pagination) ([]dto.CommonUserRelationData, int64, error) {
	var records []models.UserAchievement
	baseQuery := config.DB.Model(&models.UserAchievement{}).
		Where("user_id = ?", userID).
		Preload("Achievement")

	total, err := executePaginatedQuery(baseQuery, pagination, &records)
	if err != nil {
		return nil, 0, err
	}

	return dto.BuildCommonUserAchievementRelationList(records), total, nil
}

func queryUserSkills(userID uint, pagination models.Pagination) ([]dto.CommonUserRelationData, int64, error) {
	var records []models.UserSkill
	baseQuery := config.DB.Model(&models.UserSkill{}).
		Where("user_id = ?", userID).
		Preload("Skill")

	total, err := executePaginatedQuery(baseQuery, pagination, &records)
	if err != nil {
		return nil, 0, err
	}

	return dto.BuildCommonUserSkillRelationList(records), total, nil
}

func queryUserItems(userID uint, pagination models.Pagination) ([]dto.CommonUserRelationData, int64, error) {
	var records []models.UserItem
	baseQuery := config.DB.Model(&models.UserItem{}).
		Where("user_id = ?", userID).
		Preload("Item")

	total, err := executePaginatedQuery(baseQuery, pagination, &records)
	if err != nil {
		return nil, 0, err
	}

	return dto.BuildCommonUserItemRelationList(records), total, nil
}

func queryUserCards(userID uint, pagination models.Pagination) ([]dto.CommonUserRelationData, int64, error) {
	var records []models.UserCard
	baseQuery := config.DB.Model(&models.UserCard{}).
		Where("user_id = ?", userID).
		Preload("Card")

	total, err := executePaginatedQuery(baseQuery, pagination, &records)
	if err != nil {
		return nil, 0, err
	}

	return dto.BuildCommonUserCardRelationList(records), total, nil
}

func executePaginatedQuery(baseQuery *gorm.DB, pagination models.Pagination, dest interface{}) (int64, error) {
	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return 0, err
	}

	if err := baseQuery.Limit(pagination.Limit).Offset(pagination.Offset).Find(dest).Error; err != nil {
		return 0, err
	}

	return total, nil
}

func getRelationQueryUserID(c *gin.Context) (uint, error) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		return 0, ErrUserIDMissing
	}

	switch val := userIDValue.(type) {
	case uint:
		if val == 0 {
			return 0, ErrUserIDTypeInvalid
		}
		return val, nil
	case uint64:
		if val == 0 {
			return 0, ErrUserIDTypeInvalid
		}
		return uint(val), nil
	case int:
		if val <= 0 {
			return 0, ErrUserIDTypeInvalid
		}
		return uint(val), nil
	case int64:
		if val <= 0 {
			return 0, ErrUserIDTypeInvalid
		}
		return uint(val), nil
	default:
		return 0, ErrUserIDTypeInvalid
	}
}
