package config

import (
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	models "MuXi/2026-MuxiShooter-Backend/models"
	utils "MuXi/2026-MuxiShooter-Backend/utils"
)

const (
	NumLimter                = 20
	DefaultHeadImagePath     = "static/DefaultHeadImg.jpeg"
	PasswordUpdatedInterval  = 30 * time.Minute
	UsernameUpdatedInterval  = 24 * time.Hour
	HeadImageUpdatedInterval = 24 * time.Hour
	PrefixHeadImg            = "HeadImg"
	DefaultPage              = 1
	DefaultPageSize          = 20
	MaxPageSize              = 100
)

var (
	ErrJWTWrongSigningMethod = errors.New("无效的签名算法")
	ErrJWTSecretGenerate     = errors.New("JWT密钥生成失败")
)

type Settings struct {
	DBUser        string
	DBPassword    string
	DBHost        string
	DBPort        string
	DBName        string
	AdminUsername string
	AdminPassword string
	JWTSecret     string
}

type AppState struct {
	DB        *gorm.DB
	JWTSecret []byte
}

func LoadSettings() Settings {
	return Settings{
		DBUser:        utils.GetEnv("DB_USER", "adminuser"),
		DBPassword:    utils.GetEnv("DB_PASSWORD", ""),
		DBHost:        utils.GetEnv("DB_HOST", ""),
		DBPort:        utils.GetEnv("DB_PORT", "3306"),
		DBName:        utils.GetEnv("DB_NAME", "mini"),
		AdminUsername: utils.GetEnv("ADMIN_USERNAME", "adminuser"),
		AdminPassword: utils.GetEnv("ADMIN_PASSWORD", ""),
		JWTSecret:     utils.GetEnv("JWT_SECRET", ""),
	}
}

func Bootstrap(settings Settings) (*AppState, error) {
	db, err := connectDB(settings)
	if err != nil {
		return nil, err
	}

	if err = initAdmin(db, settings); err != nil {
		return nil, err
	}

	jwtSecret, err := initJWTSecret(settings)
	if err != nil {
		return nil, err
	}

	return &AppState{DB: db, JWTSecret: jwtSecret}, nil
}

func connectDB(settings Settings) (*gorm.DB, error) {
	if settings.DBPassword == "" {
		return nil, errors.New("数据库管理用户密码环境变量(DB_PASSWORD)为空,请配置")
	}
	if settings.DBHost == "" {
		return nil, errors.New("DB_HOST为空，请设置环境变量")
	}

	dsnRoot := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=true&loc=Local",
		settings.DBUser, settings.DBPassword, settings.DBHost, settings.DBPort)

	var rootDB *gorm.DB
	var err error
	maxRetries := 15
	for i := 0; i < maxRetries; i++ {
		rootDB, err = gorm.Open(mysql.Open(dsnRoot), &gorm.Config{})
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("连接MySQL失败: %w", err)
	}

	createDb := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4;", settings.DBName)
	if err = rootDB.Exec(createDb).Error; err != nil {
		return nil, fmt.Errorf("创建数据库失败: %w", err)
	}

	dsnDB := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		settings.DBUser, settings.DBPassword, settings.DBHost, settings.DBPort, settings.DBName)
	db, err := gorm.Open(mysql.Open(dsnDB), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	err = db.AutoMigrate(&models.Achievement{}, &models.User{}, &models.Skill{}, &models.Card{}, &models.Item{}, &models.UserAchievement{}, &models.UserCard{}, &models.UserItem{}, &models.UserSkill{})
	if err != nil {
		return nil, fmt.Errorf("数据迁移失败: %w", err)
	}

	return db, nil
}

func initAdmin(db *gorm.DB, settings Settings) error {
	admin := settings.AdminUsername
	adminPsw := settings.AdminPassword
	var count int64
	db.Model(&models.User{}).Where("username = ?", admin).Count(&count)

	if count > 0 {
		return nil
	}

	hashedPsw, err := utils.Hashtool(adminPsw)
	if err != nil {
		return fmt.Errorf("管理员密码哈希失败: %w", err)
	}

	adminUser := models.User{
		Username:      admin,
		Password:      hashedPsw,
		Group:         "admin",
		HeadImagePath: DefaultHeadImagePath,
	}

	if err := db.Create(&adminUser).Error; err != nil {
		return fmt.Errorf("初始化管理员失败: %w", err)
	}
	return nil
}

func initJWTSecret(settings Settings) ([]byte, error) {
	if len(settings.JWTSecret) == 0 {
		secret, err := utils.GenerateSercet(32)
		if err != nil {
			return nil, fmt.Errorf("%w:%v", ErrJWTSecretGenerate, err)
		}
		return secret, nil
	}

	decoded, err := base64.StdEncoding.DecodeString(settings.JWTSecret)
	if err != nil {
		return nil, fmt.Errorf("base64解码JWT密钥环境变量失败:%w", err)
	}
	if len(decoded) < 32 {
		return nil, errors.New("JWT密钥长度不足32字节")
	}
	return decoded, nil
}
