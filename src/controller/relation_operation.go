package controller

import (
	config "MuXi/2026-MuxiShooter-Backend/config"
	"MuXi/2026-MuxiShooter-Backend/dto"
	models "MuXi/2026-MuxiShooter-Backend/models"
	"errors"
	"time"

	"gorm.io/gorm"
)

var (
	ErrSkillGradeOnlyForSkills = errors.New("skill_grade仅skills类型可用")
)

type relationOperator interface {
	Create(userID uint, resourceID uint) (dto.CommonUserRelationData, error)
	Update(userID uint, req dto.UserRelationUpdateRequest) (dto.CommonUserRelationData, error)
	Delete(userID uint, resourceID uint) error
}

type relationService struct {
	operators map[UserRelationType]relationOperator
}

var selfRelationService = newRelationService(config.DB)

func newRelationService(db *gorm.DB) *relationService {
	return &relationService{
		operators: map[UserRelationType]relationOperator{
			UserRelationAchievement: &achievementRelationOperator{db: db},
			UserRelationSkill:       &skillRelationOperator{db: db},
			UserRelationItem:        &itemRelationOperator{db: db},
			UserRelationCard:        &cardRelationOperator{db: db},
		},
	}
}

func (s *relationService) getOperator(relationType UserRelationType) (relationOperator, error) {
	op, exists := s.operators[relationType]
	if !exists {
		return nil, ErrUnsupportedRelationType
	}
	return op, nil
}

func createSelfRelationByType(userID uint, relationType UserRelationType, resourceID uint) (dto.CommonUserRelationData, error) {
	op, err := selfRelationService.getOperator(relationType)
	if err != nil {
		return dto.CommonUserRelationData{}, err
	}
	return op.Create(userID, resourceID)
}

func updateSelfRelationByType(userID uint, relationType UserRelationType, req dto.UserRelationUpdateRequest) (dto.CommonUserRelationData, error) {
	op, err := selfRelationService.getOperator(relationType)
	if err != nil {
		return dto.CommonUserRelationData{}, err
	}
	return op.Update(userID, req)
}

func deleteSelfRelationByType(userID uint, relationType UserRelationType, resourceID uint) error {
	op, err := selfRelationService.getOperator(relationType)
	if err != nil {
		return err
	}
	return op.Delete(userID, resourceID)
}

type achievementRelationOperator struct {
	db *gorm.DB
}

func (op *achievementRelationOperator) Create(userID uint, resourceID uint) (dto.CommonUserRelationData, error) {
	var resource models.Achievement
	if err := op.db.First(&resource, resourceID).Error; err != nil {
		return dto.CommonUserRelationData{}, err
	}

	var existed models.UserAchievement
	err := op.db.Where("user_id = ? AND achievement_id = ?", userID, resourceID).First(&existed).Error
	if err == nil {
		return dto.CommonUserRelationData{}, ErrResourceNameExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.CommonUserRelationData{}, err
	}

	record := models.UserAchievement{UserID: userID, AchievementID: resourceID}
	if err = op.db.Create(&record).Error; err != nil {
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

type skillRelationOperator struct {
	db *gorm.DB
}

func (op *skillRelationOperator) Create(userID uint, resourceID uint) (dto.CommonUserRelationData, error) {
	var resource models.Skill
	if err := op.db.First(&resource, resourceID).Error; err != nil {
		return dto.CommonUserRelationData{}, err
	}

	var existed models.UserSkill
	err := op.db.Where("user_id = ? AND skill_id = ?", userID, resourceID).First(&existed).Error
	if err == nil {
		return dto.CommonUserRelationData{}, ErrResourceNameExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.CommonUserRelationData{}, err
	}

	record := models.UserSkill{UserID: userID, SkillID: resourceID}
	if err = op.db.Create(&record).Error; err != nil {
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

type itemRelationOperator struct {
	db *gorm.DB
}

func (op *itemRelationOperator) Create(userID uint, resourceID uint) (dto.CommonUserRelationData, error) {
	var resource models.Item
	if err := op.db.First(&resource, resourceID).Error; err != nil {
		return dto.CommonUserRelationData{}, err
	}

	var existed models.UserItem
	err := op.db.Where("user_id = ? AND item_id = ?", userID, resourceID).First(&existed).Error
	if err == nil {
		return dto.CommonUserRelationData{}, ErrResourceNameExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.CommonUserRelationData{}, err
	}

	record := models.UserItem{UserID: userID, ItemID: resourceID}
	if err = op.db.Create(&record).Error; err != nil {
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

type cardRelationOperator struct {
	db *gorm.DB
}

func (op *cardRelationOperator) Create(userID uint, resourceID uint) (dto.CommonUserRelationData, error) {
	var resource models.Card
	if err := op.db.First(&resource, resourceID).Error; err != nil {
		return dto.CommonUserRelationData{}, err
	}

	var existed models.UserCard
	err := op.db.Where("user_id = ? AND card_id = ?", userID, resourceID).First(&existed).Error
	if err == nil {
		return dto.CommonUserRelationData{}, ErrResourceNameExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.CommonUserRelationData{}, err
	}

	record := models.UserCard{UserID: userID, CardID: resourceID}
	if err = op.db.Create(&record).Error; err != nil {
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
		return nil, ErrNoUpdateFields
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
			return nil, ErrSkillGradeOnlyForSkills
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
