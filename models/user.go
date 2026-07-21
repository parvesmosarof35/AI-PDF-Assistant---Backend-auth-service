package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name              string             `bson:"name" json:"name" validate:"required"`
	Email             string             `bson:"email" json:"email" validate:"required,email"`
	Password          string             `bson:"password" json:"-" validate:"required"` // Don't return password in JSON
	IsVerified        bool               `bson:"is_verified" json:"is_verified"`
	VerificationToken string             `bson:"verification_token,omitempty" json:"-"`
	ResetToken        string             `bson:"reset_token,omitempty" json:"-"`
	ResetTokenExpiry  time.Time          `bson:"reset_token_expiry,omitempty" json:"-"`
	AvatarURL         string             `bson:"avatar_url,omitempty" json:"avatar_url,omitempty"`
	CreatedAt         time.Time          `bson:"created_at" json:"created_at"`
}

type SignupInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type VerifyEmailInput struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required,len=6"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UpdateProfileInput struct {
	Name      string `json:"name,omitempty"`
	Email     string `json:"email,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

type ChangePasswordInput struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type ForgotPasswordInput struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordInput struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}
