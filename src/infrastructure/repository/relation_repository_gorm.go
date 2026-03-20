package repository

import (
	"MuXi/2026-MuxiShooter-Backend/dto"
	"MuXi/2026-MuxiShooter-Backend/models"
	"MuXi/2026-MuxiShooter-Backend/service"
	"errors"
	"time"

	"gorm.io/gorm"
)

type RelationRepositoryGorm struct {
	db *gorm.DB
}

func NewRelationRepository(db *gorm.DB) *RelationRepositoryGorm {
	return &RelationRepositoryGorm{db: db}
}

func (r *RelationRepositoryGorm) CreateUserRelation(userID uint, relationType service.UserRelationType, resourceID uint) (dto.CommonUserRelationData, error) {
	operator, err := r.getOperator(relationType)
	if err != nil {
		return dto.CommonUserRelationData{}, err
	}
	return operator.Create(userID, resourceID)
}

func (r *RelationRepositoryGorm) UpdateUserRelation(userID uint, relationType service.UserRelationType, req dto.UserRelationUpdateRequest) (dto.CommonUserRelationData, error) {
	operator, err := r.getOperator(relationType)
	if err != nil {
		return dto.CommonUserRelationData{}, err
	}
	return operator.Update(userID, req)
}

func (r *RelationRepositoryGorm) DeleteUserRelation(userID uint, relationType service.UserRelationType, resourceID uint) error {
	operator, err := r.getOperator(relationType)
	if err != nil {
		return err
	}
	return operator.Delete(userID, resourceID)
}

func (r *RelationRepositoryGorm) QueryUserRelationsByType(userID uint, relationType service.UserRelationType, pagination models.Pagination) ([]dto.CommonUserRelationData, int64, error) {
	switch relationType {
	case service.UserRelationAchievement:
		var records []models.UserAchievement
		baseQuery := r.db.Model(&models.UserAchievement{}).Where("user_id = ?", userID).Preload("Achievement")
		total, err := executePaginatedQuery(baseQuery, pagination, &records)
		if err != nil {
			return nil, 0, err
		}
		return dto.BuildCommonUserAchievementRelationList(records), total, nil
	case service.UserRelationSkill:
		var records []models.UserSkill
		baseQuery := r.db.Model(&models.UserSkill{}).Where("user_id = ?", userID).Preload("Skill")
		total, err := executePaginatedQuery(baseQuery, pagination, &records)
		if err != nil {
			return nil, 0, err
		}
		return dto.BuildCommonUserSkillRelationList(records), total, nil
	case service.UserRelationItem:
		var records []models.UserItem
		baseQuery := r.db.Model(&models.UserItem{}).Where("user_id = ?", userID).Preload("Item")
		total, err := executePaginatedQuery(baseQuery, pagination, &records)
		if err != nil {
			return nil, 0, err
		}
		return dto.BuildCommonUserItemRelationList(records), total, nil
	case service.UserRelationCard:
		var records []models.UserCard
		baseQuery := r.db.Model(&models.UserCard{}).Where("user_id = ?", userID).Preload("Card")
		total, err := executePaginatedQuery(baseQuery, pagination, &records)
		if err != nil {
			return nil, 0, err
		}
		return dto.BuildCommonUserCardRelationList(records), total, nil
	default:
		return nil, 0, service.ErrUnsupportedRelationType
	}
}

type relationOperator interface {
	Create(userID uint, resourceID uint) (dto.CommonUserRelationData, error)
	Update(userID uint, req dto.UserRelationUpdateRequest) (dto.CommonUserRelationData, error)
	Delete(userID uint, resourceID uint) error
}

func (r *RelationRepositoryGorm) getOperator(relationType service.UserRelationType) (relationOperator, error) {
	switch relationType {
	case service.UserRelationAchievement:
		return &achievementRelationOperator{db: r.db}, nil
	case service.UserRelationSkill:
		return &skillRelationOperator{db: r.db}, nil
	case service.UserRelationItem:
		return &itemRelationOperator{db: r.db}, nil
	case service.UserRelationCard:
		return &cardRelationOperator{db: r.db}, nil
	default:
		return nil, service.ErrUnsupportedRelationType
	}
}

type achievementRelationOperator struct{ db *gorm.DB }

func (op *achievementRelationOperator) Create(userID uint, resourceID uint) (dto.CommonUserRelationData, error) {
	var resource models.Achievement
	if err := op.db.First(&resource, resourceID).Error; err != nil {
		return dto.CommonUserRelationData{}, err
	}

	var existed models.UserAchievement
	err := op.db.Where("user_id = ? AND achievement_id = ?", userID, resourceID).First(&existed).Error
	if err == nil {
		return dto.CommonUserRelationData{}, service.ErrResourceNameExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.CommonUserRelationData{}, err
	}

	record := models.UserAchievement{UserID: userID, AchievementID: resourceID}
	if err = createAndEnsureOneRow(op.db, &record); err != nil {
		return dto.CommonUserRelationData{}, err
	}
	if err = op.db.Preload("Achievement").Where("user_id = ? AND achievement_id = ?", userID, resourceID).First(&record).Error; err != nil {
		return dto.CommonUserRelationData{}, err
	}

	return dto.BuildCommonUserAchievementRelationList([]models.UserAchievement{record})[0], nil
}

func (op *achievementRelationOperator) Update(userID uint, req dto.UserRelationUpdateRequest) (dto.CommonUserRelationData, error) {
	updates, err := buildRelationStatusUpdates(req, false)
	if err != nil {
		return dto.CommonUserRelationData{}, err
	}

	var record models.UserAchievement
	if err = op.db.Where("user_id = ? AND achievement_id = ?", userID, req.ResourceID).First(&record).Error; err != nil {
		return dto.CommonUserRelationData{}, err
	}
	if err = op.db.Model(&record).Updates(updates).Error; err != nil {
		return dto.CommonUserRelationData{}, err
	}
	if err = op.db.Preload("Achievement").Where("user_id = ? AND achievement_id = ?", userID, req.ResourceID).First(&record).Error; err != nil {
		return dto.CommonUserRelationData{}, err
	}

	return dto.BuildCommonUserAchievementRelationList([]models.UserAchievement{record})[0], nil
}

func (op *achievementRelationOperator) Delete(userID uint, resourceID uint) error {
	result := op.db.Where("user_id = ? AND achievement_id = ?", userID, resourceID).Delete(&models.UserAchievement{})
	return mapDeleteResult(result)
}

type skillRelationOperator struct{ db *gorm.DB }

func (op *skillRelationOperator) Create(userID uint, resourceID uint) (dto.CommonUserRelationData, error) {
	var resource models.Skill
	if err := op.db.First(&resource, resourceID).Error; err != nil {
		return dto.CommonUserRelationData{}, err
	}

	var existed models.UserSkill
	err := op.db.Where("user_id = ? AND skill_id = ?", userID, resourceID).First(&existed).Error
	if err == nil {
		return dto.CommonUserRelationData{}, service.ErrResourceNameExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.CommonUserRelationData{}, err
	}

	record := models.UserSkill{UserID: userID, SkillID: resourceID}
	if err = createAndEnsureOneRow(op.db, &record); err != nil {
		return dto.CommonUserRelationData{}, err
	}
	if err = op.db.Preload("Skill").Where("user_id = ? AND skill_id = ?", userID, resourceID).First(&record).Error; err != nil {
		return dto.CommonUserRelationData{}, err
	}

	return dto.BuildCommonUserSkillRelationList([]models.UserSkill{record})[0], nil
}

func (op *skillRelationOperator) Update(userID uint, req dto.UserRelationUpdateRequest) (dto.CommonUserRelationData, error) {
	updates, err := buildRelationStatusUpdates(req, true)
	if err != nil {
		return dto.CommonUserRelationData{}, err
	}

	var record models.UserSkill
	if err = op.db.Where("user_id = ? AND skill_id = ?", userID, req.ResourceID).First(&record).Error; err != nil {
		return dto.CommonUserRelationData{}, err
	}
	if err = op.db.Model(&record).Updates(updates).Error; err != nil {
		return dto.CommonUserRelationData{}, err
	}
	if err = op.db.Preload("Skill").Where("user_id = ? AND skill_id = ?", userID, req.ResourceID).First(&record).Error; err != nil {
		return dto.CommonUserRelationData{}, err
	}

	return dto.BuildCommonUserSkillRelationList([]models.UserSkill{record})[0], nil
}

func (op *skillRelationOperator) Delete(userID uint, resourceID uint) error {
	result := op.db.Where("user_id = ? AND skill_id = ?", userID, resourceID).Delete(&models.UserSkill{})
	return mapDeleteResult(result)
}

type itemRelationOperator struct{ db *gorm.DB }

func (op *itemRelationOperator) Create(userID uint, resourceID uint) (dto.CommonUserRelationData, error) {
	var resource models.Item
	if err := op.db.First(&resource, resourceID).Error; err != nil {
		return dto.CommonUserRelationData{}, err
	}

	var existed models.UserItem
	err := op.db.Where("user_id = ? AND item_id = ?", userID, resourceID).First(&existed).Error
	if err == nil {
		return dto.CommonUserRelationData{}, service.ErrResourceNameExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.CommonUserRelationData{}, err
	}

	record := models.UserItem{UserID: userID, ItemID: resourceID}
	if err = createAndEnsureOneRow(op.db, &record); err != nil {
		return dto.CommonUserRelationData{}, err
	}
	if err = op.db.Preload("Item").Where("user_id = ? AND item_id = ?", userID, resourceID).First(&record).Error; err != nil {
		return dto.CommonUserRelationData{}, err
	}

	return dto.BuildCommonUserItemRelationList([]models.UserItem{record})[0], nil
}

func (op *itemRelationOperator) Update(userID uint, req dto.UserRelationUpdateRequest) (dto.CommonUserRelationData, error) {
	updates, err := buildRelationStatusUpdates(req, false)
	if err != nil {
		return dto.CommonUserRelationData{}, err
	}

	var record models.UserItem
	if err = op.db.Where("user_id = ? AND item_id = ?", userID, req.ResourceID).First(&record).Error; err != nil {
		return dto.CommonUserRelationData{}, err
	}
	if err = op.db.Model(&record).Updates(updates).Error; err != nil {
		return dto.CommonUserRelationData{}, err
	}
	if err = op.db.Preload("Item").Where("user_id = ? AND item_id = ?", userID, req.ResourceID).First(&record).Error; err != nil {
		return dto.CommonUserRelationData{}, err
	}

	return dto.BuildCommonUserItemRelationList([]models.UserItem{record})[0], nil
}

func (op *itemRelationOperator) Delete(userID uint, resourceID uint) error {
	result := op.db.Where("user_id = ? AND item_id = ?", userID, resourceID).Delete(&models.UserItem{})
	return mapDeleteResult(result)
}

type cardRelationOperator struct{ db *gorm.DB }

func (op *cardRelationOperator) Create(userID uint, resourceID uint) (dto.CommonUserRelationData, error) {
	var resource models.Card
	if err := op.db.First(&resource, resourceID).Error; err != nil {
		return dto.CommonUserRelationData{}, err
	}

	var existed models.UserCard
	err := op.db.Where("user_id = ? AND card_id = ?", userID, resourceID).First(&existed).Error
	if err == nil {
		return dto.CommonUserRelationData{}, service.ErrResourceNameExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.CommonUserRelationData{}, err
	}

	record := models.UserCard{UserID: userID, CardID: resourceID}
	if err = createAndEnsureOneRow(op.db, &record); err != nil {
		return dto.CommonUserRelationData{}, err
	}
	if err = op.db.Preload("Card").Where("user_id = ? AND card_id = ?", userID, resourceID).First(&record).Error; err != nil {
		return dto.CommonUserRelationData{}, err
	}

	return dto.BuildCommonUserCardRelationList([]models.UserCard{record})[0], nil
}

func (op *cardRelationOperator) Update(userID uint, req dto.UserRelationUpdateRequest) (dto.CommonUserRelationData, error) {
	updates, err := buildRelationStatusUpdates(req, false)
	if err != nil {
		return dto.CommonUserRelationData{}, err
	}

	var record models.UserCard
	if err = op.db.Where("user_id = ? AND card_id = ?", userID, req.ResourceID).First(&record).Error; err != nil {
		return dto.CommonUserRelationData{}, err
	}
	if err = op.db.Model(&record).Updates(updates).Error; err != nil {
		return dto.CommonUserRelationData{}, err
	}
	if err = op.db.Preload("Card").Where("user_id = ? AND card_id = ?", userID, req.ResourceID).First(&record).Error; err != nil {
		return dto.CommonUserRelationData{}, err
	}

	return dto.BuildCommonUserCardRelationList([]models.UserCard{record})[0], nil
}

func (op *cardRelationOperator) Delete(userID uint, resourceID uint) error {
	result := op.db.Where("user_id = ? AND card_id = ?", userID, resourceID).Delete(&models.UserCard{})
	return mapDeleteResult(result)
}

func buildRelationStatusUpdates(req dto.UserRelationUpdateRequest, allowSkillGrade bool) (map[string]interface{}, error) {
	if req.IsComplete == nil && req.Claimed == nil && req.SkillGrade == nil {
		return nil, service.ErrNoUpdateFields
	}

	now := time.Now()
	updates := map[string]interface{}{}
	if req.IsComplete != nil {
		updates["is_complete"] = *req.IsComplete
		if *req.IsComplete {
			updates["complete_at"] = &now
		} else {
			updates["complete_at"] = nil
		}
	}
	if req.Claimed != nil {
		updates["claimed"] = *req.Claimed
		if *req.Claimed {
			updates["claimed_at"] = &now
		} else {
			updates["claimed_at"] = nil
		}
	}
	if req.SkillGrade != nil {
		if !allowSkillGrade {
			return nil, service.ErrSkillGradeOnlyForSkills
		}
		updates["skill_grade"] = *req.SkillGrade
	}

	return updates, nil
}

func mapDeleteResult(result *gorm.DB) error {
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
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

func createAndEnsureOneRow(db *gorm.DB, model interface{}) error {
	result := db.Create(model)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return service.ErrRelationCreateNoRows
	}
	return nil
}
