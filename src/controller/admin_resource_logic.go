package controller

import (
	"MuXi/2026-MuxiShooter-Backend/dto"
	models "MuXi/2026-MuxiShooter-Backend/models"
	utils "MuXi/2026-MuxiShooter-Backend/utils"
	"errors"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var (
	ErrResourceTypeRequired = errors.New("缺少type参数")
	ErrResourceIDInvalid    = errors.New("id参数格式错误")
	ErrNoUpdateFields       = errors.New("没有可更新字段")
	ErrResourceNameExists   = errors.New("同类型资源名称已存在")
	ErrInvalidRequestBody   = errors.New("请求参数错误")
)

type adminResourceQueryHandler func(c *gin.Context, pagination models.Pagination) ([]dto.CommonAdminResourceData, int64, error)
type adminResourceCreateHandler func(c *gin.Context) (dto.CommonAdminResourceData, error)
type adminResourceUpdateHandler func(c *gin.Context) (dto.CommonAdminResourceData, error)
type adminResourceDeleteHandler func(id uint) error

var adminResourceQueryHandlers = map[UserRelationType]adminResourceQueryHandler{
	UserRelationAchievement: adminQueryAchievements,
	UserRelationSkill:       adminQuerySkills,
	UserRelationItem:        adminQueryItems,
	UserRelationCard:        adminQueryCards,
}

var adminResourceCreateHandlers = map[UserRelationType]adminResourceCreateHandler{
	UserRelationAchievement: adminCreateAchievement,
	UserRelationSkill:       adminCreateSkill,
	UserRelationItem:        adminCreateItem,
	UserRelationCard:        adminCreateCard,
}

var adminResourceUpdateHandlers = map[UserRelationType]adminResourceUpdateHandler{
	UserRelationAchievement: adminUpdateAchievement,
	UserRelationSkill:       adminUpdateSkill,
	UserRelationItem:        adminUpdateItem,
	UserRelationCard:        adminUpdateCard,
}

var adminResourceDeleteHandlers = map[UserRelationType]adminResourceDeleteHandler{
	UserRelationAchievement: adminDeleteAchievement,
	UserRelationSkill:       adminDeleteSkill,
	UserRelationItem:        adminDeleteItem,
	UserRelationCard:        adminDeleteCard,
}

func parseResourceType(c *gin.Context) (UserRelationType, error) {
	typeStr := c.Query("type")
	if typeStr == "" {
		return "", ErrResourceTypeRequired
	}
	return ParseUserRelationType(typeStr)
}

func parseOptionalID(c *gin.Context) (uint, bool, error) {
	idStr := c.Query("id")
	if idStr == "" {
		return 0, false, nil
	}
	idValue, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || idValue == 0 {
		return 0, false, ErrResourceIDInvalid
	}
	return uint(idValue), true, nil
}

func adminQueryAchievements(c *gin.Context, pagination models.Pagination) ([]dto.CommonAdminResourceData, int64, error) {
	id, hasID, err := parseOptionalID(c)
	if err != nil {
		return nil, 0, err
	}
	name := utils.SqlSafeLikeKeyword(c.Query("name"))

	if hasID {
		var record models.Achievement
		err = currentDB().First(&record, id).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []dto.CommonAdminResourceData{}, 0, nil
		}
		if err != nil {
			return nil, 0, err
		}
		return []dto.CommonAdminResourceData{dto.BuildCommonAdminAchievementData(record)}, 1, nil
	}

	db := currentDB().Model(&models.Achievement{})
	if name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
	}
	var total int64
	if err = db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []models.Achievement
	if err = db.Limit(pagination.Limit).Offset(pagination.Offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return dto.BuildCommonAdminAchievementList(list), total, nil
}

func adminQuerySkills(c *gin.Context, pagination models.Pagination) ([]dto.CommonAdminResourceData, int64, error) {
	id, hasID, err := parseOptionalID(c)
	if err != nil {
		return nil, 0, err
	}
	name := utils.SqlSafeLikeKeyword(c.Query("name"))
	skillGroup := utils.SqlSafeLikeKeyword(c.Query("skill_group"))

	if hasID {
		var record models.Skill
		err = currentDB().First(&record, id).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []dto.CommonAdminResourceData{}, 0, nil
		}
		if err != nil {
			return nil, 0, err
		}
		return []dto.CommonAdminResourceData{dto.BuildCommonAdminSkillData(record)}, 1, nil
	}

	db := currentDB().Model(&models.Skill{})
	if name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
	}
	if skillGroup != "" {
		db = db.Where("skill_group LIKE ?", "%"+skillGroup+"%")
	}
	var total int64
	if err = db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []models.Skill
	if err = db.Limit(pagination.Limit).Offset(pagination.Offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return dto.BuildCommonAdminSkillList(list), total, nil
}

func adminQueryItems(c *gin.Context, pagination models.Pagination) ([]dto.CommonAdminResourceData, int64, error) {
	id, hasID, err := parseOptionalID(c)
	if err != nil {
		return nil, 0, err
	}
	name := utils.SqlSafeLikeKeyword(c.Query("name"))

	if hasID {
		var record models.Item
		err = currentDB().First(&record, id).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []dto.CommonAdminResourceData{}, 0, nil
		}
		if err != nil {
			return nil, 0, err
		}
		return []dto.CommonAdminResourceData{dto.BuildCommonAdminItemData(record)}, 1, nil
	}

	db := currentDB().Model(&models.Item{})
	if name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
	}
	var total int64
	if err = db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []models.Item
	if err = db.Limit(pagination.Limit).Offset(pagination.Offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return dto.BuildCommonAdminItemList(list), total, nil
}

func adminQueryCards(c *gin.Context, pagination models.Pagination) ([]dto.CommonAdminResourceData, int64, error) {
	id, hasID, err := parseOptionalID(c)
	if err != nil {
		return nil, 0, err
	}
	name := utils.SqlSafeLikeKeyword(c.Query("name"))

	if hasID {
		var record models.Card
		err = currentDB().First(&record, id).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []dto.CommonAdminResourceData{}, 0, nil
		}
		if err != nil {
			return nil, 0, err
		}
		return []dto.CommonAdminResourceData{dto.BuildCommonAdminCardData(record)}, 1, nil
	}

	db := currentDB().Model(&models.Card{})
	if name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
	}
	var total int64
	if err = db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []models.Card
	if err = db.Limit(pagination.Limit).Offset(pagination.Offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return dto.BuildCommonAdminCardList(list), total, nil
}

func adminCreateAchievement(c *gin.Context) (dto.CommonAdminResourceData, error) {
	var req dto.AdminCreateAchievementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return dto.CommonAdminResourceData{}, fmt.Errorf("%w:%v", ErrInvalidRequestBody, err)
	}
	if err := ensureUniqueResourceName(UserRelationAchievement, req.Name, nil); err != nil {
		return dto.CommonAdminResourceData{}, err
	}
	record := models.Achievement{Name: req.Name, Description: req.Description}
	if err := currentDB().Create(&record).Error; err != nil {
		return dto.CommonAdminResourceData{}, err
	}
	return dto.BuildCommonAdminAchievementData(record), nil
}

func adminCreateSkill(c *gin.Context) (dto.CommonAdminResourceData, error) {
	var req dto.AdminCreateSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return dto.CommonAdminResourceData{}, fmt.Errorf("%w:%v", ErrInvalidRequestBody, err)
	}
	if err := ensureUniqueResourceName(UserRelationSkill, req.Name, nil); err != nil {
		return dto.CommonAdminResourceData{}, err
	}
	record := models.Skill{Name: req.Name, Description: req.Description, SkillGroup: req.SkillGroup, PrqSkillId: req.PrqSkillID}
	if err := currentDB().Create(&record).Error; err != nil {
		return dto.CommonAdminResourceData{}, err
	}
	return dto.BuildCommonAdminSkillData(record), nil
}

func adminCreateItem(c *gin.Context) (dto.CommonAdminResourceData, error) {
	var req dto.AdminCreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return dto.CommonAdminResourceData{}, fmt.Errorf("%w:%v", ErrInvalidRequestBody, err)
	}
	if err := ensureUniqueResourceName(UserRelationItem, req.Name, nil); err != nil {
		return dto.CommonAdminResourceData{}, err
	}
	record := models.Item{Name: req.Name, Description: req.Description}
	if err := currentDB().Create(&record).Error; err != nil {
		return dto.CommonAdminResourceData{}, err
	}
	return dto.BuildCommonAdminItemData(record), nil
}

func adminCreateCard(c *gin.Context) (dto.CommonAdminResourceData, error) {
	var req dto.AdminCreateCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return dto.CommonAdminResourceData{}, fmt.Errorf("%w:%v", ErrInvalidRequestBody, err)
	}
	if err := ensureUniqueResourceName(UserRelationCard, req.Name, nil); err != nil {
		return dto.CommonAdminResourceData{}, err
	}
	record := models.Card{Name: req.Name, Description: req.Description}
	if err := currentDB().Create(&record).Error; err != nil {
		return dto.CommonAdminResourceData{}, err
	}
	return dto.BuildCommonAdminCardData(record), nil
}

func adminUpdateAchievement(c *gin.Context) (dto.CommonAdminResourceData, error) {
	var req dto.AdminUpdateAchievementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return dto.CommonAdminResourceData{}, fmt.Errorf("%w:%v", ErrInvalidRequestBody, err)
	}
	updates := map[string]interface{}{}
	if req.Name != nil {
		if err := ensureUniqueResourceName(UserRelationAchievement, *req.Name, &req.ID); err != nil {
			return dto.CommonAdminResourceData{}, err
		}
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if len(updates) == 0 {
		return dto.CommonAdminResourceData{}, ErrNoUpdateFields
	}
	return doUpdateAchievement(req.ID, updates)
}

func adminUpdateSkill(c *gin.Context) (dto.CommonAdminResourceData, error) {
	var req dto.AdminUpdateSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return dto.CommonAdminResourceData{}, fmt.Errorf("%w:%v", ErrInvalidRequestBody, err)
	}
	updates := map[string]interface{}{}
	if req.Name != nil {
		if err := ensureUniqueResourceName(UserRelationSkill, *req.Name, &req.ID); err != nil {
			return dto.CommonAdminResourceData{}, err
		}
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.SkillGroup != nil {
		updates["skill_group"] = *req.SkillGroup
	}
	if req.PrqSkillID != nil {
		updates["prq_skill_id"] = *req.PrqSkillID
	}
	if len(updates) == 0 {
		return dto.CommonAdminResourceData{}, ErrNoUpdateFields
	}
	return doUpdateSkill(req.ID, updates)
}

func adminUpdateItem(c *gin.Context) (dto.CommonAdminResourceData, error) {
	var req dto.AdminUpdateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return dto.CommonAdminResourceData{}, fmt.Errorf("%w:%v", ErrInvalidRequestBody, err)
	}
	updates := map[string]interface{}{}
	if req.Name != nil {
		if err := ensureUniqueResourceName(UserRelationItem, *req.Name, &req.ID); err != nil {
			return dto.CommonAdminResourceData{}, err
		}
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if len(updates) == 0 {
		return dto.CommonAdminResourceData{}, ErrNoUpdateFields
	}
	return doUpdateItem(req.ID, updates)
}

func adminUpdateCard(c *gin.Context) (dto.CommonAdminResourceData, error) {
	var req dto.AdminUpdateCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return dto.CommonAdminResourceData{}, fmt.Errorf("%w:%v", ErrInvalidRequestBody, err)
	}
	updates := map[string]interface{}{}
	if req.Name != nil {
		if err := ensureUniqueResourceName(UserRelationCard, *req.Name, &req.ID); err != nil {
			return dto.CommonAdminResourceData{}, err
		}
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if len(updates) == 0 {
		return dto.CommonAdminResourceData{}, ErrNoUpdateFields
	}
	return doUpdateCard(req.ID, updates)
}

func doUpdateAchievement(id uint, updates map[string]interface{}) (dto.CommonAdminResourceData, error) {
	var record models.Achievement
	if err := currentDB().First(&record, id).Error; err != nil {
		return dto.CommonAdminResourceData{}, err
	}
	if err := currentDB().Model(&record).Updates(updates).Error; err != nil {
		return dto.CommonAdminResourceData{}, err
	}
	if err := currentDB().First(&record, id).Error; err != nil {
		return dto.CommonAdminResourceData{}, err
	}
	return dto.BuildCommonAdminAchievementData(record), nil
}

func doUpdateSkill(id uint, updates map[string]interface{}) (dto.CommonAdminResourceData, error) {
	var record models.Skill
	if err := currentDB().First(&record, id).Error; err != nil {
		return dto.CommonAdminResourceData{}, err
	}
	if err := currentDB().Model(&record).Updates(updates).Error; err != nil {
		return dto.CommonAdminResourceData{}, err
	}
	if err := currentDB().First(&record, id).Error; err != nil {
		return dto.CommonAdminResourceData{}, err
	}
	return dto.BuildCommonAdminSkillData(record), nil
}

func doUpdateItem(id uint, updates map[string]interface{}) (dto.CommonAdminResourceData, error) {
	var record models.Item
	if err := currentDB().First(&record, id).Error; err != nil {
		return dto.CommonAdminResourceData{}, err
	}
	if err := currentDB().Model(&record).Updates(updates).Error; err != nil {
		return dto.CommonAdminResourceData{}, err
	}
	if err := currentDB().First(&record, id).Error; err != nil {
		return dto.CommonAdminResourceData{}, err
	}
	return dto.BuildCommonAdminItemData(record), nil
}

func doUpdateCard(id uint, updates map[string]interface{}) (dto.CommonAdminResourceData, error) {
	var record models.Card
	if err := currentDB().First(&record, id).Error; err != nil {
		return dto.CommonAdminResourceData{}, err
	}
	if err := currentDB().Model(&record).Updates(updates).Error; err != nil {
		return dto.CommonAdminResourceData{}, err
	}
	if err := currentDB().First(&record, id).Error; err != nil {
		return dto.CommonAdminResourceData{}, err
	}
	return dto.BuildCommonAdminCardData(record), nil
}

func adminDeleteAchievement(id uint) error {
	return doDeleteByID(&models.Achievement{}, id)
}

func adminDeleteSkill(id uint) error {
	return doDeleteByID(&models.Skill{}, id)
}

func adminDeleteItem(id uint) error {
	return doDeleteByID(&models.Item{}, id)
}

func adminDeleteCard(id uint) error {
	return doDeleteByID(&models.Card{}, id)
}

func ensureUniqueResourceName(resourceType UserRelationType, name string, excludeID *uint) error {
	if name == "" {
		return nil
	}

	var count int64
	queryByModel := func(model interface{}) error {
		db := currentDB().Model(model).Where("name = ?", name)
		if excludeID != nil && *excludeID > 0 {
			db = db.Where("id <> ?", *excludeID)
		}
		return db.Count(&count).Error
	}

	var err error
	switch resourceType {
	case UserRelationAchievement:
		err = queryByModel(&models.Achievement{})
	case UserRelationSkill:
		err = queryByModel(&models.Skill{})
	case UserRelationItem:
		err = queryByModel(&models.Item{})
	case UserRelationCard:
		err = queryByModel(&models.Card{})
	default:
		return ErrUnsupportedRelationType
	}
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrResourceNameExists
	}
	return nil
}

func doDeleteByID(model interface{}, id uint) error {
	result := currentDB().Delete(model, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
