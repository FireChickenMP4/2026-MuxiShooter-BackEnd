package dto

import (
	models "MuXi/2026-MuxiShooter-Backend/models"
	"time"
)

type AdminAchievementData struct {
	AchievementID   uint      `json:"achievement_id"`
	AchievementName string    `json:"achievement_name"`
	Description     string    `json:"description"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type AdminSkillData struct {
	SkillID     uint      `json:"skill_id"`
	SkillName   string    `json:"skill_name"`
	Description string    `json:"description"`
	SkillGroup  string    `json:"skill_group"`
	PrqSkillID  uint      `json:"prq_skill_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type AdminItemData struct {
	ItemID      uint      `json:"item_id"`
	ItemName    string    `json:"item_name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type AdminCardData struct {
	CardID      uint      `json:"card_id"`
	CardName    string    `json:"card_name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type AdminAchievementPageData struct {
	List     []AdminAchievementData `json:"list"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
}

type AdminSkillPageData struct {
	List     []AdminSkillData `json:"list"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

type AdminItemPageData struct {
	List     []AdminItemData `json:"list"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

type AdminCardPageData struct {
	List     []AdminCardData `json:"list"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

func BuildAdminAchievementData(record models.Achievement) AdminAchievementData {
	return AdminAchievementData{
		AchievementID:   record.ID,
		AchievementName: record.Name,
		Description:     record.Description,
		CreatedAt:       record.CreatedAt,
		UpdatedAt:       record.UpdatedAt,
	}
}

func BuildAdminSkillData(record models.Skill) AdminSkillData {
	return AdminSkillData{
		SkillID:     record.ID,
		SkillName:   record.Name,
		Description: record.Description,
		SkillGroup:  record.SkillGroup,
		PrqSkillID:  record.PrqSkillId,
		CreatedAt:   record.CreatedAt,
		UpdatedAt:   record.UpdatedAt,
	}
}

func BuildAdminItemData(record models.Item) AdminItemData {
	return AdminItemData{
		ItemID:      record.ID,
		ItemName:    record.Name,
		Description: record.Description,
		CreatedAt:   record.CreatedAt,
		UpdatedAt:   record.UpdatedAt,
	}
}

func BuildAdminCardData(record models.Card) AdminCardData {
	return AdminCardData{
		CardID:      record.ID,
		CardName:    record.Name,
		Description: record.Description,
		CreatedAt:   record.CreatedAt,
		UpdatedAt:   record.UpdatedAt,
	}
}

func BuildAdminAchievementList(records []models.Achievement) []AdminAchievementData {
	result := make([]AdminAchievementData, 0, len(records))
	for _, record := range records {
		result = append(result, BuildAdminAchievementData(record))
	}
	return result
}

func BuildAdminSkillList(records []models.Skill) []AdminSkillData {
	result := make([]AdminSkillData, 0, len(records))
	for _, record := range records {
		result = append(result, BuildAdminSkillData(record))
	}
	return result
}

func BuildAdminItemList(records []models.Item) []AdminItemData {
	result := make([]AdminItemData, 0, len(records))
	for _, record := range records {
		result = append(result, BuildAdminItemData(record))
	}
	return result
}

func BuildAdminCardList(records []models.Card) []AdminCardData {
	result := make([]AdminCardData, 0, len(records))
	for _, record := range records {
		result = append(result, BuildAdminCardData(record))
	}
	return result
}
