package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/time_capsule/Auth-Servic-Timecapsule/config"
	"github.com/time_capsule/Auth-Servic-Timecapsule/internal/auth"
	"github.com/time_capsule/Auth-Servic-Timecapsule/internal/email"
	"github.com/time_capsule/Auth-Servic-Timecapsule/internal/models"
	"github.com/time_capsule/Auth-Servic-Timecapsule/internal/redis"
	"github.com/time_capsule/Auth-Servic-Timecapsule/internal/user"
)

// AuthHandler handles authentication-related API requests.
type AuthHandler struct {
	userRepo    *user.UserRepo
	redisClient *redis.Client
	cfg         *config.Config
	jwtManager  *auth.JWTManager
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(db *pgxpool.Pool, redisClient *redis.Client, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		userRepo:    user.NewUserRepo(db),
		redisClient: redisClient,
		cfg:         cfg,
		jwtManager:  auth.NewJWTManager(cfg),
	}
}

// Register godoc
// @Summary      Register a new user
// @Description  Registers a new user and sends an OTP to their email for verification.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user  body     models.UserCreate  true  "User registration data"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      409  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var input models.User
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	_, err := h.userRepo.GetUserByEmail(context.Background(), input.Email)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User with this email already exists"})
		return
	}

	// Hash the password
	hashedPassword, err := auth.HashPassword(input.PasswordHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	input.PasswordHash = hashedPassword

	// Generate OTP
	otp := auth.GenerateOTP()

	// Save OTP in Redis
	err = h.redisClient.SaveOTP(context.Background(), input.Email, otp, 5*time.Minute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save OTP"})
		return
	}

	// Send OTP email
	err = email.SendOTP(h.cfg, input.Email, otp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP email" + err.Error()})
		return
	}

	// Create the user (without saving the password yet)
	input.PasswordHash = "" // Don't save the password until OTP is verified
	if err := h.userRepo.CreateUser(context.Background(), &input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully. Please verify your email."})
}

// VerifyOTP godoc
// @Summary      Verify OTP
// @Description  Verifies the OTP sent to the user's email during registration.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body      VerifyOTPInput  true  "Email and OTP"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /auth/verify-otp [post]
func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		OTP      string `json:"otp"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify OTP against Redis
	isValid, err := h.redisClient.VerifyOTP(context.Background(), input.Email, input.OTP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify OTP"})
		return
	}
	if !isValid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid OTP"})
		return
	}

	// Update the user with the hashed password
	hashedPassword, err := auth.HashPassword(input.Password) // Use OTP as the password for now
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	if err := h.userRepo.UpdateUserPassword(context.Background(), &models.UserUpdatePass{Email: input.Email, PasswordHash: hashedPassword}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP verified successfully"})
}

// Login godoc
// @Summary      Login
// @Description  Authenticates a user and issues a JWT token.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body      LoginInput  true  "User login credentials"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the user from the database
	user, err := h.userRepo.GetUserByEmail(context.Background(), input.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}
	if user.Role == "courier" && user.Status != "approved" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Your account should have been approved."})
		return
	}
	// Compare the provided password with the stored hash
	if !auth.CheckPasswordHash(input.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Generate JWT token
	token, err := h.jwtManager.Generate(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// Validate godoc
// @Summary      Validate Token
// @Description  Validates a JWT token and returns the user ID and role.
// @Tags         auth
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /auth/validate [get]
func (h *AuthHandler) Validate(c *gin.Context) {
	// Get the token from the Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		return
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Verify the token
	claims, err := h.jwtManager.Verify(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":   claims.GetUserID(),
		"role": claims.GetUserRole(),
	})
}

// ForgotPassword godoc
// @Summary      Forgot Password
// @Description  Initiates the password reset process by sending an OTP to the user's email.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body      ForgotPasswordInput  true  "User's email address"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var input struct {
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error while binding email" + err.Error()})
		return
	}

	// Check if user exists
	_, err := h.userRepo.GetUserByEmail(context.Background(), input.Email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Generate OTP
	otp := auth.GenerateOTP()

	// Save OTP in Redis (with a longer expiration, e.g., 15 minutes)
	err = h.redisClient.SaveOTP(context.Background(), input.Email, otp, 15*time.Minute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save OTP"})
		return
	}

	// Send OTP email
	err = email.SendOTP(h.cfg, input.Email, otp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset OTP sent to your email."})
}

// ResetPassword godoc
// @Summary      Reset Password
// @Description  Resets the user's password using the provided OTP and new password.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body      ResetPasswordInput  true  "Email, OTP, and new password"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var input struct {
		Email              string `json:"email" binding:"required,email"`
		OTP                string `json:"otp" binding:"required"`
		NewPassword        string `json:"new_password" binding:"required"`
		ConfirmNewPassword string `json:"confirm_new_password" binding:"required,eqfield=NewPassword"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify OTP against Redis
	isValid, err := h.redisClient.VerifyOTP(context.Background(), input.Email, input.OTP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify OTP"})
		return
	}
	if !isValid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid OTP"})
		return
	}

	// Hash the new password
	hashedPassword, err := auth.HashPassword(input.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Update the user's password in the database
	if err := h.userRepo.UpdateUserPassword(context.Background(), &models.UserUpdatePass{Email: input.Email, PasswordHash: hashedPassword}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

// ApproveUser 	 godoc
// @Summary      Approve users.
// @Description  Approve users account for requesting courier.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body    models.UserUpdateStatus  true  "Email and status"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /auth/approve-user [post]
func (h *AuthHandler) ApproveUser(c *gin.Context) {
	var req models.UserUpdateStatus
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.userRepo.UpdateUserStatus(context.Background(), &models.UserUpdateStatus{Email: req.Email, Status: req.Status}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Status updated successfully"})
}

// Input Structs
type ForgotPasswordInput struct {
	Email string `json:"email"`
}

type ResetPasswordInput struct {
	Email              string `json:"email" binding:"required,email"`
	OTP                string `json:"otp" binding:"required"`
	NewPassword        string `json:"new_password" binding:"required"`
	ConfirmNewPassword string `json:"confirm_new_password" binding:"required,eqfield=NewPassword"`
}

// LoginInput represents the input for the login endpoint.
type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// VerifyOTPInput represents the input for the OTP verification endpoint.
type VerifyOTPInput struct {
	Email    string `json:"email" binding:"required,email"`
	OTP      string `json:"otp" binding:"required"`
	Password string `json:"password" binding:"required"`
}
