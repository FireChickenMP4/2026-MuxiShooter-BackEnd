package controller

import (
	"errors"

	"gorm.io/gorm"
)

var (
	ErrControllerDBNotInitialized        = errors.New("controller db is not initialized")
	ErrControllerJWTSecretNotInitialized = errors.New("controller jwt secret is not initialized")
)

var (
	appDB        *gorm.DB
	appJWTSecret []byte
)

func SetDB(db *gorm.DB) {
	appDB = db
}

func SetJWTSecret(secret []byte) {
	appJWTSecret = secret
}

func ValidateDependencies() error {
	if appDB == nil {
		return ErrControllerDBNotInitialized
	}
	if len(appJWTSecret) == 0 {
		return ErrControllerJWTSecretNotInitialized
	}
	return nil
}

func currentDB() *gorm.DB {
	if appDB == nil {
		panic(ErrControllerDBNotInitialized.Error())
	}
	return appDB
}

func currentJWTSecret() []byte {
	if len(appJWTSecret) == 0 {
		panic(ErrControllerJWTSecretNotInitialized.Error())
	}
	return appJWTSecret
}
