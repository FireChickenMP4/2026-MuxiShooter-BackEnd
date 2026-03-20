package controller

import (
	"MuXi/2026-MuxiShooter-Backend/dto"
	models "MuXi/2026-MuxiShooter-Backend/models"
	"errors"

	"gorm.io/gorm"
)

var (
	ErrUnsupportedRelationType = errors.New("不支持的关联表类型")
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
	baseQuery := currentDB().Model(&models.UserAchievement{}).
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
	baseQuery := currentDB().Model(&models.UserSkill{}).
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
	baseQuery := currentDB().Model(&models.UserItem{}).
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
	baseQuery := currentDB().Model(&models.UserCard{}).
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

