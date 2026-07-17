package users

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/spdeepak/aegis/server/api"
	"github.com/spdeepak/aegis/server/internal/db"
)

func TestAdminService_LockUserById_OK(t *testing.T) {
	truncateTables()
	dbConnection := db.Connect(dbConfig)
	userQuery := New(dbConnection)
	admin_service := NewAdminService(userQuery)
	signupNo2faOk(t)
	email, err := userQuery.GetUserByEmail(context.Background(), "first.last@example.com")
	assert.NoError(t, err)
	ctx := context.WithValue(context.Background(), "User-ID", email.ID)
	ctx = context.WithValue(ctx, "user-ip", "127.0.0.1")
	users, err := userQuery.GetUserByEmail(ctx, "first.last@example.com")
	assert.NoError(t, err)
	assert.NotEmpty(t, users)
	assert.False(t, users.Locked)
	lockUser(t, ctx, err, admin_service, users, userQuery)
	dbConnection.Close()
}

func TestAdminService_LockUserById_NOK(t *testing.T) {
	truncateTables()
	dbConnection := db.Connect(dbConfig)
	userQuery := New(dbConnection)
	admin_service := NewAdminService(userQuery)
	signupNo2faOk(t)
	email, err := userQuery.GetUserByEmail(context.Background(), "first.last@example.com")
	assert.NoError(t, err)
	ctx := context.WithValue(context.Background(), "User-ID", email.ID)
	ctx = context.WithValue(ctx, "user-ip", "127.0.0.1")
	err = admin_service.LockUserById(ctx, int64(9999999), api.LockUserParams{UserAgent: "service-test"})
	assert.Error(t, err)
	dbConnection.Close()
}

func TestAdminService_UnlockUserById_OK(t *testing.T) {
	truncateTables()
	dbConnection := db.Connect(dbConfig)
	userQuery := New(dbConnection)
	admin_service := NewAdminService(userQuery)
	signupNo2faOk(t)
	email, err := userQuery.GetUserByEmail(context.Background(), "first.last@example.com")
	assert.NoError(t, err)
	ctx := context.WithValue(context.Background(), "User-ID", email.ID)
	ctx = context.WithValue(ctx, "user-ip", "127.0.0.1")
	users, err := userQuery.GetUserByEmail(ctx, "first.last@example.com")
	assert.NoError(t, err)
	assert.NotEmpty(t, users)
	assert.False(t, users.Locked)
	lockUser(t, ctx, err, admin_service, users, userQuery)
	unlockUser(t, err, admin_service, users, userQuery)
	dbConnection.Close()
}

func TestAdminService_UnlockUserById_NOK(t *testing.T) {
	truncateTables()
	dbConnection := db.Connect(dbConfig)
	userQuery := New(dbConnection)
	admin_service := NewAdminService(userQuery)
	ctx := context.WithValue(context.Background(), "User-ID", int64(999999))
	ctx = context.WithValue(ctx, "user-ip", "127.0.0.1")
	err := admin_service.UnlockUserById(ctx, int64(9999999), api.UnlockUserParams{UserAgent: "service-test"})
	assert.Error(t, err)
	dbConnection.Close()
}

func lockUser(t *testing.T, ctx context.Context, err error, admin_service AdminService, users User, userQuery *Queries) {
	err = admin_service.LockUserById(ctx, users.ID, api.LockUserParams{UserAgent: "service-test"})
	assert.NoError(t, err)
	users, err = userQuery.GetUserByEmail(ctx, "first.last@example.com")
	assert.NoError(t, err)
	assert.NotEmpty(t, users)
	assert.True(t, users.Locked)
}

func unlockUser(t *testing.T, err error, admin_service AdminService, users User, userQuery *Queries) {
	ctx := context.WithValue(context.Background(), "User-ID", int64(99999999))
	ctx = context.WithValue(ctx, "user-ip", "127.0.0.1")
	err = admin_service.UnlockUserById(ctx, users.ID, api.UnlockUserParams{UserAgent: "service-test"})
	assert.NoError(t, err)
	users, err = userQuery.GetUserByEmail(ctx, "first.last@example.com")
	assert.NoError(t, err)
	assert.NotEmpty(t, users)
	assert.False(t, users.Locked)
}

func TestAdminService_DisableUserById_OK(t *testing.T) {
	truncateTables()
	dbConnection := db.Connect(dbConfig)
	userQuery := New(dbConnection)
	admin_service := NewAdminService(userQuery)
	signupNo2faOk(t)
	email, err := userQuery.GetUserByEmail(context.Background(), "first.last@example.com")
	assert.NoError(t, err)
	ctx := context.WithValue(context.Background(), "User-ID", email.ID)
	ctx = context.WithValue(ctx, "user-ip", "127.0.0.1")
	users, err := userQuery.GetUserByEmail(ctx, "first.last@example.com")
	assert.NoError(t, err)
	assert.NotEmpty(t, users)
	assert.False(t, users.Disabled)
	disableUser(t, err, admin_service, users, userQuery)
	dbConnection.Close()
}

func TestAdminService_DisableUserById_NOK(t *testing.T) {
	truncateTables()
	dbConnection := db.Connect(dbConfig)
	userQuery := New(dbConnection)
	admin_service := NewAdminService(userQuery)
	ctx := context.WithValue(context.Background(), "User-ID", int64(999999))
	ctx = context.WithValue(ctx, "user-ip", "127.0.0.1")
	err := admin_service.DisableUserById(ctx, int64(99999999), api.DisableUserParams{UserAgent: "service-test"})
	assert.Error(t, err)
	dbConnection.Close()
}

func TestAdminService_EnableUserById_OK(t *testing.T) {
	truncateTables()
	dbConnection := db.Connect(dbConfig)
	userQuery := New(dbConnection)
	admin_service := NewAdminService(userQuery)
	signupNo2faOk(t)
	email, err := userQuery.GetUserByEmail(context.Background(), "first.last@example.com")
	assert.NoError(t, err)
	ctx := context.WithValue(context.Background(), "User-ID", email.ID)
	ctx = context.WithValue(ctx, "user-ip", "127.0.0.1")
	users, err := userQuery.GetUserByEmail(ctx, "first.last@example.com")
	assert.NoError(t, err)
	assert.NotEmpty(t, users)
	assert.False(t, users.Locked)
	disableUser(t, err, admin_service, users, userQuery)
	enableUser(t, err, admin_service, users, userQuery)
	dbConnection.Close()
}

func TestAdminService_EnableUserById_NOK(t *testing.T) {
	truncateTables()
	dbConnection := db.Connect(dbConfig)
	userQuery := New(dbConnection)
	admin_service := NewAdminService(userQuery)
	ctx := context.WithValue(context.Background(), "User-ID", int64(999999))
	ctx = context.WithValue(ctx, "user-ip", "127.0.0.1")
	err := admin_service.EnableUserById(ctx, int64(99999998), api.EnableUserParams{UserAgent: "service-test"})
	assert.Error(t, err)
	dbConnection.Close()
}

func disableUser(t *testing.T, err error, admin_service AdminService, users User, userQuery *Queries) {
	ctx := context.WithValue(context.Background(), "User-ID", int64(99999999))
	ctx = context.WithValue(ctx, "user-ip", "127.0.0.1")
	err = admin_service.DisableUserById(ctx, users.ID, api.DisableUserParams{UserAgent: "service-test"})
	assert.NoError(t, err)
	users, err = userQuery.GetUserByEmail(ctx, "first.last@example.com")
	assert.NoError(t, err)
	assert.NotEmpty(t, users)
	assert.True(t, users.Disabled)
}

func enableUser(t *testing.T, err error, admin_service AdminService, users User, userQuery *Queries) {
	ctx := context.WithValue(context.Background(), "User-ID", int64(99999999))
	ctx = context.WithValue(ctx, "user-ip", "127.0.0.1")
	err = admin_service.EnableUserById(ctx, users.ID, api.EnableUserParams{UserAgent: "service-test"})
	assert.NoError(t, err)
	users, err = userQuery.GetUserByEmail(ctx, "first.last@example.com")
	assert.NoError(t, err)
	assert.NotEmpty(t, users)
	assert.False(t, users.Disabled)
}
