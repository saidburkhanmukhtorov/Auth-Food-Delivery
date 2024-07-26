package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/time_capsule/Auth-Servic-Timecapsule/internal/models"
	"github.com/time_capsule/Auth-Servic-Timecapsule/internal/user"
)

// UserHandler handles user-related API requests.
type UserHandler struct {
	userRepo *user.UserRepo
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(db *pgxpool.Pool) *UserHandler {
	return &UserHandler{
		userRepo: user.NewUserRepo(db),
	}
}

// GetAllUsers godoc
// @Summary      Get All Users
// @Description  Retrieves a list of all users.
// @Tags         users
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        email  query      string  false  "Email"
// @Param 		 fullname query string false "Fullname"
// @Param 		 username query string false "Username"
// @Param 		 status query string false "Status"
// @Param 		 role query string false "Role"
// @Success      200  {array}   models.User
// @Failure      500  {object}  map[string]interface{}
// @Router       /users [get]
func (h *UserHandler) GetAllUsers(c *gin.Context) {
	var (
		userReq models.GetAllUsers
	)
	userReq.Email = c.Query("email")
	userReq.FullName = c.Query("fullname")
	userReq.Username = c.Query("username")
	userReq.Status = c.Query("status")
	userReq.Role = c.Query("role")

	users, err := h.userRepo.GetAllUsers(context.Background(), userReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get users"})
		return
	}

	c.JSON(http.StatusOK, users)
}

// GetUserByID godoc
// @Summary      Get User by ID
// @Description  Retrieves a user by their ID.
// @Tags         users
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        userId  path      string  true  "User ID"
// @Success      200  {object}  models.User
// @Failure      400  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /users/{userId} [get]
func (h *UserHandler) GetUserByID(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.userRepo.GetUserByID(context.Background(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateUser godoc
// @Summary      Update User
// @Description  Updates a user's information.
// @Tags         users
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        userId  path      string  true  "User ID"
// @Param        user  body      models.User  true  "Updated user data"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /users/{userId} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("userId")

	var input models.UserUpdate
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = userID // Ensure the ID is set correctly

	if err := h.userRepo.UpdateUser(context.Background(), &input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

// DeleteUser godoc
// @Summary      Delete User
// @Description  Deletes a user by their ID.
// @Tags         users
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        userId  path      string  true  "User ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /users/{userId} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := h.userRepo.DeleteUser(context.Background(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
