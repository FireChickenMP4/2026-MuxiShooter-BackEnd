package service

import (
	config "MuXi/2026-MuxiShooter-Backend/config"
	"MuXi/2026-MuxiShooter-Backend/dto"
	"MuXi/2026-MuxiShooter-Backend/models"
	"errors"
	"time"
)

var (
	ErrMissingUserContext         = errors.New("解析后token中缺少用户信息")
	ErrPasswordTooFrequent        = errors.New("修改密码间隔过短")
	ErrUsernameTooFrequent        = errors.New("修改用户名间隔过短")
	ErrHeadImageTooFrequent       = errors.New("修改头像间隔过短")
	ErrSamePassword               = errors.New("所给新旧密码不能相同")
	ErrInvalidOldPassword         = errors.New("旧密码错误")
	ErrMissingCoinType            = errors.New("缺少type参数")
	ErrUnsupportedCoinType        = errors.New("type参数仅支持strength/select")
	ErrUnsupportedRelationType    = errors.New("不支持的关联表类型")
	ErrMissingRelationType        = errors.New("缺少type参数")
	ErrNoUpdateFields             = errors.New("没有可更新字段")
	ErrResourceNameExists         = errors.New("同类型资源名称已存在")
	ErrSkillGradeOnlyForSkills    = errors.New("skill_grade仅skills类型可用")
	ErrRelationCreateNoRows       = errors.New("创建关联失败：数据库未写入任何记录")
	ErrRelationCreateInconsistent = errors.New("创建关联失败：返回数据与请求不一致")
)

type UserRelationType string

const (
	UserRelationAchievement UserRelationType = "achievements"
	UserRelationSkill       UserRelationType = "skills"
	UserRelationItem        UserRelationType = "items"
	UserRelationCard        UserRelationType = "cards"
)

func ParseUserRelationType(val string) (UserRelationType, error) {
	relationType := UserRelationType(val)
	switch relationType {
	case UserRelationAchievement, UserRelationSkill, UserRelationItem, UserRelationCard:
		return relationType, nil
	default:
		return "", ErrUnsupportedRelationType
	}
}

type ProfileUserRepository interface {
	FindByID(userID uint) (*models.User, bool, error)
	UpdatePassword(userID uint, hashedPassword string, updatedAt time.Time) error
	UpdateUsername(userID uint, newUsername string, updatedAt time.Time) error
	UpdateHeadImage(userID uint, newHeadImagePath string, updatedAt time.Time) error
	UpdateCoinByField(userID uint, field string, coin uint) error
	IncrementTokenVersion(userID uint) error
}

type ProfileRelationRepository interface {
	CreateUserRelation(userID uint, relationType UserRelationType, resourceID uint) (dto.CommonUserRelationData, error)
	UpdateUserRelation(userID uint, relationType UserRelationType, req dto.UserRelationUpdateRequest) (dto.CommonUserRelationData, error)
	DeleteUserRelation(userID uint, relationType UserRelationType, resourceID uint) error
	QueryUserRelationsByType(userID uint, relationType UserRelationType, pagination models.Pagination) ([]dto.CommonUserRelationData, int64, error)
}

type ProfileService struct {
	userRepository     ProfileUserRepository
	relationRepository ProfileRelationRepository
	passwordHasher     PasswordHasher
}

func NewProfileService(userRepository ProfileUserRepository, relationRepository ProfileRelationRepository, passwordHasher PasswordHasher) *ProfileService {
	return &ProfileService{
		userRepository:     userRepository,
		relationRepository: relationRepository,
		passwordHasher:     passwordHasher,
	}
}

func (s *ProfileService) Logout(userID uint) error {
	_, existed, err := s.userRepository.FindByID(userID)
	if err != nil {
		return err
	}
	if !existed {
		return ErrUserNotFound
	}
	if err = s.userRepository.IncrementTokenVersion(userID); err != nil {
		return err
	}
	return nil
}

func (s *ProfileService) UpdatePassword(userID uint, req dto.UpdatePasswordRequest) error {
	if req.NewPassword == req.OldPassword {
		return ErrSamePassword
	}

	user, existed, err := s.userRepository.FindByID(userID)
	if err != nil {
		return err
	}
	if !existed || user == nil {
		return ErrUserNotFound
	}

	if user.PasswordUpdatedAt != nil && time.Since(*user.PasswordUpdatedAt) <= config.PasswordUpdatedInterval {
		return ErrPasswordTooFrequent
	}

	if err = s.passwordHasher.Compare(user.Password, req.OldPassword); err != nil {
		return ErrInvalidOldPassword
	}

	hashedPassword, err := s.passwordHasher.Hash(req.NewPassword)
	if err != nil {
		return err
	}

	now := time.Now()
	if err = s.userRepository.UpdatePassword(userID, hashedPassword, now); err != nil {
		return err
	}

	_ = s.userRepository.IncrementTokenVersion(userID)
	return nil
}

func (s *ProfileService) UpdateUsername(userID uint, req dto.UpdateUsernameRequest) error {
	user, existed, err := s.userRepository.FindByID(userID)
	if err != nil {
		return err
	}
	if !existed || user == nil {
		return ErrUserNotFound
	}

	if user.UsernameUpdatedAt != nil && time.Since(*user.UsernameUpdatedAt) <= config.UsernameUpdatedInterval {
		return ErrUsernameTooFrequent
	}

	now := time.Now()
	if err = s.userRepository.UpdateUsername(userID, req.NewUsername, now); err != nil {
		return err
	}

	return nil
}

func (s *ProfileService) UpdateHeadImage(userID uint, newHeadImagePath string) (string, error) {
	user, existed, err := s.userRepository.FindByID(userID)
	if err != nil {
		return "", err
	}
	if !existed || user == nil {
		return "", ErrUserNotFound
	}

	if user.HeadImageUpdatedAt != nil && time.Since(*user.HeadImageUpdatedAt) <= config.HeadImageUpdatedInterval {
		return "", ErrHeadImageTooFrequent
	}

	now := time.Now()
	if err = s.userRepository.UpdateHeadImage(userID, newHeadImagePath, now); err != nil {
		return "", err
	}

	return user.HeadImagePath, nil
}

func (s *ProfileService) UpdateCoinByType(userID uint, coinType string, coin uint) (dto.CommonUserData, error) {
	if coinType == "" {
		return dto.CommonUserData{}, ErrMissingCoinType
	}

	updateField := ""
	switch coinType {
	case "strength", "strength_coin":
		updateField = "strength_coin"
	case "select", "select_coin":
		updateField = "select_coin"
	default:
		return dto.CommonUserData{}, ErrUnsupportedCoinType
	}

	user, existed, err := s.userRepository.FindByID(userID)
	if err != nil {
		return dto.CommonUserData{}, err
	}
	if !existed || user == nil {
		return dto.CommonUserData{}, ErrUserNotFound
	}

	if err = s.userRepository.UpdateCoinByField(userID, updateField, coin); err != nil {
		return dto.CommonUserData{}, err
	}

	if updateField == "strength_coin" {
		user.StrengthCoin = coin
	} else {
		user.SelectCoin = coin
	}

	return dto.CommonUserData{
		UserID:        user.ID,
		Username:      user.Username,
		Group:         user.Group,
		HeadImagePath: user.HeadImagePath,
		StrengthCoin:  user.StrengthCoin,
		SelectCoin:    user.SelectCoin,
	}, nil
}

func (s *ProfileService) CreateSelfRelationByType(userID uint, relationType UserRelationType, resourceID uint) (dto.CommonUserRelationData, error) {
	return s.relationRepository.CreateUserRelation(userID, relationType, resourceID)
}

func (s *ProfileService) UpdateSelfRelationByType(userID uint, relationType UserRelationType, req dto.UserRelationUpdateRequest) (dto.CommonUserRelationData, error) {
	return s.relationRepository.UpdateUserRelation(userID, relationType, req)
}

func (s *ProfileService) DeleteSelfRelationByType(userID uint, relationType UserRelationType, resourceID uint) error {
	return s.relationRepository.DeleteUserRelation(userID, relationType, resourceID)
}

func (s *ProfileService) GetSelfProfile(userID uint) (dto.CommonUserData, error) {
	user, existed, err := s.userRepository.FindByID(userID)
	if err != nil {
		return dto.CommonUserData{}, err
	}
	if !existed || user == nil {
		return dto.CommonUserData{}, ErrUserNotFound
	}

	return dto.CommonUserData{
		UserID:        user.ID,
		Username:      user.Username,
		Group:         user.Group,
		HeadImagePath: user.HeadImagePath,
		StrengthCoin:  user.StrengthCoin,
		SelectCoin:    user.SelectCoin,
	}, nil
}

func (s *ProfileService) GetSelfRelationsByType(userID uint, relationTypeStr string, pagination models.Pagination) ([]dto.CommonUserRelationData, int64, error) {
	if relationTypeStr == "" {
		return nil, 0, ErrMissingRelationType
	}
	relationType, err := ParseUserRelationType(relationTypeStr)
	if err != nil {
		return nil, 0, err
	}

	return s.relationRepository.QueryUserRelationsByType(userID, relationType, pagination)
}
