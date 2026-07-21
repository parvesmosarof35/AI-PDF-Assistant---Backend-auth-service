package controllers

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"auth-service/config"
	"auth-service/models"
	"auth-service/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"github.com/google/uuid"
)

func Signup(c *gin.Context) {
	var input models.SignupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userCollection := config.DB.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if user already exists
	var existingUser models.User
	err := userCollection.FindOne(ctx, bson.M{"email": input.Email}).Decode(&existingUser)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User with this email already exists"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Generate 6-digit verification code
	verificationCode := fmt.Sprintf("%06d", rand.Intn(1000000))

	// Create user
	newUser := models.User{
		Name:              input.Name,
		Email:             input.Email,
		Password:          string(hashedPassword),
		IsVerified:        false,
		VerificationToken: verificationCode,
		CreatedAt:         time.Now(),
	}

	result, err := userCollection.InsertOne(ctx, newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	userID := result.InsertedID.(primitive.ObjectID).Hex()

	// Send Verification Email asynchronously
	go func() {
		_ = utils.SendVerificationEmail(newUser.Email, verificationCode)
	}()

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully. Please check your email for the 6-digit verification code.",
		"needs_verification": true,
		"user": gin.H{
			"id":          userID,
			"name":        newUser.Name,
			"email":       newUser.Email,
			"is_verified": newUser.IsVerified,
		},
	})
}

func Login(c *gin.Context) {
	var input models.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userCollection := config.DB.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User
	err := userCollection.FindOne(ctx, bson.M{"email": input.Email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	if !user.IsVerified {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Please verify your email first",
			"needs_verification": true,
			"email": user.Email,
		})
		return
	}

	// Generate token
	token, err := utils.GenerateToken(user.ID.Hex(), user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
		"user": gin.H{
			"id":          user.ID.Hex(),
			"name":        user.Name,
			"email":       user.Email,
			"is_verified": user.IsVerified,
			"avatar_url":  user.AvatarURL,
		},
	})
}

func VerifyEmail(c *gin.Context) {
	var input models.VerifyEmailInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userCollection := config.DB.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find the user by email
	var user models.User
	err := userCollection.FindOne(ctx, bson.M{"email": input.Email}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}

	if user.IsVerified {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is already verified"})
		return
	}

	if user.VerificationToken != input.Code {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid verification code"})
		return
	}

	// Update user
	update := bson.M{
		"$set": bson.M{"is_verified": true},
		"$unset": bson.M{"verification_token": ""}, // Remove token after successful verification
	}

	_, err = userCollection.UpdateOne(ctx, bson.M{"_id": user.ID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email successfully verified"})
}

func ForgotPassword(c *gin.Context) {
	var input models.ForgotPasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userCollection := config.DB.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if user exists
	var user models.User
	err := userCollection.FindOne(ctx, bson.M{"email": input.Email}).Decode(&user)
	if err != nil {
		// Don't leak whether the email exists or not
		c.JSON(http.StatusOK, gin.H{"message": "If that email exists, a reset link has been sent."})
		return
	}

	// Generate reset token
	resetToken := uuid.New().String()
	resetExpiry := time.Now().Add(15 * time.Minute)

	update := bson.M{
		"$set": bson.M{
			"reset_token":        resetToken,
			"reset_token_expiry": resetExpiry,
		},
	}

	_, err = userCollection.UpdateOne(ctx, bson.M{"_id": user.ID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Send Email asynchronously
	go func() {
		_ = utils.SendPasswordResetEmail(user.Email, resetToken)
	}()

	c.JSON(http.StatusOK, gin.H{"message": "If that email exists, a reset link has been sent."})
}

func ResetPassword(c *gin.Context) {
	var input models.ResetPasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userCollection := config.DB.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find user by valid reset token
	var user models.User
	err := userCollection.FindOne(ctx, bson.M{
		"reset_token": input.Token,
		"reset_token_expiry": bson.M{"$gt": time.Now()},
	}).Decode(&user)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired reset token"})
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), 12)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Update password and remove reset token
	update := bson.M{
		"$set": bson.M{"password": string(hashedPassword)},
		"$unset": bson.M{
			"reset_token":        "",
			"reset_token_expiry": "",
		},
	}

	_, err = userCollection.UpdateOne(ctx, bson.M{"_id": user.ID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password successfully reset. You can now login."})
}
