package repository

import (
	"MuXi/2026-MuxiShooter-Backend/models"
	"errors"
	"time"

	"gorm.io/gorm"
)

type UserRepositoryGorm struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepositoryGorm {
	return &UserRepositoryGorm{db: db}
}

func (r *UserRepositoryGorm) FindByUsername(username string) (*models.User, bool, error) {
	var user models.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return &user, true, nil
}

func (r *UserRepositoryGorm) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepositoryGorm) FindByID(userID uint) (*models.User, bool, error) {
	var user models.User
	err := r.db.First(&user, userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return &user, true, nil
}

func (r *UserRepositoryGorm) UpdatePassword(userID uint, hashedPassword string, updatedAt time.Time) error {
	result := r.db.Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"password":            hashedPassword,
			"password_updated_at": updatedAt,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *UserRepositoryGorm) UpdateUsername(userID uint, newUsername string, updatedAt time.Time) error {
	result := r.db.Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"username":            newUsername,
			"username_updated_at": updatedAt,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *UserRepositoryGorm) UpdateHeadImage(userID uint, newHeadImagePath string, updatedAt time.Time) error {
	result := r.db.Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"head_image_path":       newHeadImagePath,
			"head_image_updated_at": updatedAt,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *UserRepositoryGorm) UpdateCoinByField(userID uint, field string, coin uint) error {
	result := r.db.Model(&models.User{}).
		Where("id = ?", userID).
		Update(field, coin)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *UserRepositoryGorm) IncrementTokenVersion(userID uint) error {
	result := r.db.Model(&models.User{}).
		Where("id = ?", userID).
		UpdateColumn("token_version", gorm.Expr("token_version + 1"))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
