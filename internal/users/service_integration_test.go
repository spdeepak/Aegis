package users

import (
	"context"
	"errors"
	"fmt"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spdeepak/go-jwt-server/api"
	"github.com/spdeepak/go-jwt-server/config"
	"github.com/spdeepak/go-jwt-server/internal/db"
	"github.com/spdeepak/go-jwt-server/internal/error"
	"github.com/spdeepak/go-jwt-server/internal/permissions"
	"github.com/spdeepak/go-jwt-server/internal/roles"
	"github.com/spdeepak/go-jwt-server/internal/tokens"
	"github.com/spdeepak/go-jwt-server/internal/twoFA"
)

var dbConfig = config.PostgresConfig{
	Host:              "localhost",
	Port:              "5432",
	DBName:            "jwt_server",
	UserName:          "admin",
	Password:          "admin",
	SSLMode:           "disable",
	Timeout:           5 * time.Second,
	MaxRetry:          5,
	ConnectTimeout:    10 * time.Second,
	StatementTimeout:  15 * time.Second,
	MaxOpenConns:      4,
	MaxIdleConns:      4,
	ConnMaxLifetime:   10 * time.Minute,
	ConnMaxIdleTime:   2 * time.Minute,
	HealthCheckPeriod: 1 * time.Minute,
}

func TestMain(m *testing.M) {
	dbConnection := db.Connect(dbConfig)
	// Run all tests
	truncateTables()
	code := m.Run()
	// Optional: Clean up (e.g., drop DB or close connection)
	truncateTables()
	dbConnection.Close()
	os.Exit(code)
}

func truncateTables() {
	t := &testing.T{}
	dbConnection := db.Connect(dbConfig)
	defer dbConnection.Close()
	_, err := dbConnection.Exec(context.Background(), `
        DO $$
			DECLARE
				r RECORD;
			BEGIN
				FOR r IN
					SELECT tablename
					FROM pg_tables
					WHERE schemaname = 'public'
				LOOP
					EXECUTE format('TRUNCATE TABLE public.%I CASCADE;', r.tablename);
				END LOOP;
		END $$;
    `)
	require.NoError(t, err)
}

func TestIntegrationService_Signup_No2FA(t *testing.T) {
	truncateTables()
	t.Run("Create New User without 2FA", func(t *testing.T) {
		signupNo2faOk(t)
	})
	t.Run("Create User already exists without 2FA", func(t *testing.T) {
		signupNo2faNokUseralreadyexists(t)
	})
}

func signupNo2faOk(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	user := api.UserSignup{
		Email:     "first.last@example.com",
		FirstName: "First name",
		LastName:  "Last name",
		Password:  "Som€_$trong_P@$$word",
	}
	dbConnection := db.Connect(dbConfig)
	defer dbConnection.Close()
	userStorage := New(dbConnection)
	userService := NewService(userStorage, nil, nil)

	res, err := userService.Signup(ctx, user)
	assert.NoError(t, err)
	assert.Empty(t, res)
}

func signupNo2faNokUseralreadyexists(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	user := api.UserSignup{
		Email:     "first.last@example.com",
		FirstName: "First name",
		LastName:  "Last name",
		Password:  "Som€_$trong_P@$$word",
	}

	dbConnection := db.Connect(dbConfig)
	defer dbConnection.Close()
	userStorage := New(dbConnection)
	userService := NewService(userStorage, nil, nil)

	res, err := userService.Signup(ctx, user)
	assert.Error(t, err)
	var he httperror.HttpError
	assert.True(t, errors.As(err, &he))
	assert.Equal(t, httperror.UserAlreadyExists, he.ErrorCode)
	assert.Empty(t, res)
}

func TestIntegrationService_Signup_2FA(t *testing.T) {
	truncateTables()
	t.Run("Create New User with 2FA", func(t *testing.T) {
		signup2faOk(t)
	})
	t.Run("Create User already exists with 2FA", func(t *testing.T) {
		signup2faNokUseralreadyexists(t)
	})
}

func signup2faOk(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	user := api.UserSignup{
		Email:        "first.last@example.com",
		FirstName:    "First name",
		LastName:     "Last name",
		Password:     "Som€_$trong_P@$$word",
		TwoFAEnabled: true,
	}

	twoFaService := twoFA.NewService("go-jwt-server", nil)
	dbConnection := db.Connect(dbConfig)
	defer dbConnection.Close()
	userStorage := New(dbConnection)
	userService := NewService(userStorage, twoFaService, nil)

	res, err := userService.Signup(ctx, user)
	assert.NoError(t, err)
	assert.NotEmpty(t, res)
	assert.NotEmpty(t, res.Secret)
	assert.NotEmpty(t, res.QrImage)
}

func signup2faNokUseralreadyexists(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	user := api.UserSignup{
		Email:        "first.last@example.com",
		FirstName:    "First name",
		LastName:     "Last name",
		Password:     "Som€_$trong_P@$$word",
		TwoFAEnabled: true,
	}

	twoFaService := twoFA.NewService("go-jwt-server", nil)
	dbConnection := db.Connect(dbConfig)
	defer dbConnection.Close()
	userStorage := New(dbConnection)
	userService := NewService(userStorage, twoFaService, nil)

	res, err := userService.Signup(ctx, user)
	assert.Error(t, err)
	var he httperror.HttpError
	assert.True(t, errors.As(err, &he))
	assert.Equal(t, httperror.UserAlreadyExists, he.ErrorCode)
	assert.Empty(t, res)
}

func TestIntegrationService_Login_OK(t *testing.T) {
	truncateTables()
	t.Run("Login without 2FA", func(t *testing.T) {
		signupNo2faOk(t)
		loginOk(t)
	})
	t.Run("Login with 2FA invalid password", func(t *testing.T) {
		loginNokWrongPassword(t)
	})
	truncateTables()
	t.Run("Login with 2FA invalid user", func(t *testing.T) {
		loginNOK(t)
	})
}

func loginOk(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Header("x-login-source", "test")
	ctx.Header("User-Agent", "test")
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.100")
	ctx.Request = req

	email := "first.last@example.com"
	userLogin := api.UserLogin{
		Email:    email,
		Password: "Som€_$trong_P@$$word",
	}

	dbConnection := db.Connect(dbConfig)
	defer dbConnection.Close()
	tokenQuery := tokens.New(dbConnection)
	tokenService := tokens.NewService(tokenQuery, []byte("JWT_$€Cr€t"), "")
	userStorage := New(dbConnection)
	userService := NewService(userStorage, nil, tokenService)
	loginParams := api.LoginParams{
		XLoginSource: api.LoginParamsXLoginSourceApi,
		UserAgent:    "test",
	}
	res, err := userService.Login(ctx, loginParams, userLogin)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotEmpty(t, res.(api.LoginSuccessWithJWT).AccessToken)
	assert.NotEmpty(t, res.(api.LoginSuccessWithJWT).RefreshToken)
}

func loginNokWrongPassword(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	email := "first.last@example.com"
	userLogin := api.UserLogin{
		Email:    email,
		Password: "Som€_P@$$word",
	}

	dbConnection := db.Connect(dbConfig)
	defer dbConnection.Close()
	userStorage := New(dbConnection)
	userService := NewService(userStorage, nil, nil)

	loginParams := api.LoginParams{
		XLoginSource: api.LoginParamsXLoginSourceApi,
		UserAgent:    "test",
	}
	res, err := userService.Login(ctx, loginParams, userLogin)
	assert.Error(t, err)
	assert.NotNil(t, res)
	assert.Empty(t, res.(api.LoginSuccessWithJWT).AccessToken)
	assert.Empty(t, res.(api.LoginSuccessWithJWT).RefreshToken)
}

func loginNOK(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	email := "first.last@example.com"
	userLogin := api.UserLogin{
		Email:    email,
		Password: "Som€_$trong_P@$$word",
	}

	dbConnection := db.Connect(dbConfig)
	defer dbConnection.Close()
	userStorage := New(dbConnection)
	userService := NewService(userStorage, nil, nil)
	loginParams := api.LoginParams{
		XLoginSource: api.LoginParamsXLoginSourceApi,
		UserAgent:    "test",
	}
	res, err := userService.Login(ctx, loginParams, userLogin)
	assert.Error(t, err)
	assert.NotNil(t, res)
	assert.Empty(t, res.(api.LoginSuccessWithJWT).AccessToken)
	assert.Empty(t, res.(api.LoginSuccessWithJWT).RefreshToken)
}

func TestIntegrationService_Login2FA(t *testing.T) {
	truncateTables()
	t.Run("Login with 2FA", func(t *testing.T) {
		login2FAOK(t)
	})
	truncateTables()
	t.Run("Login with expired 2FA", func(t *testing.T) {
		login2faNOKOld2FACode(t)
	})
	truncateTables()
	t.Run("Login with expired 2FA", func(t *testing.T) {
		login2FANOKUserNotExist(t)
	})
}

func login2FAOK(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	user := api.UserSignup{
		Email:        "first.last@example.com",
		FirstName:    "First name",
		LastName:     "Last name",
		Password:     "Som€_$trong_P@$$word",
		TwoFAEnabled: true,
	}

	secret := "JWT_$€CR€T"
	dbConnection := db.Connect(dbConfig)
	defer dbConnection.Close()
	tokenQuery := tokens.New(dbConnection)
	tokenService := tokens.NewService(tokenQuery, []byte(secret), "")
	twoFAQuery := twoFA.New(dbConnection)
	twoFaService := twoFA.NewService("go-jwt-server", twoFAQuery)
	userStorage := New(dbConnection)
	userService := NewService(userStorage, twoFaService, tokenService)

	res, err := userService.Signup(ctx, user)
	assert.NoError(t, err)
	assert.NotEmpty(t, res)
	assert.NotEmpty(t, res.Secret)
	assert.NotEmpty(t, res.QrImage)

	ctx.Header("x-login-source", "test")
	ctx.Header("user-agent", "test")
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.100")
	ctx.Request = req

	userByEmail, err := userStorage.GetEntireUserByEmail(context.Background(), "first.last@example.com")
	assert.NoError(t, err)

	passcode, err := totp.GenerateCode(res.Secret, time.Now().Add(-20*time.Second))
	assert.NoError(t, err)

	login2FA, err := userService.Login2FA(ctx, api.Login2FAParams{}, userByEmail.UserID, passcode)
	assert.NoError(t, err)
	assert.NotEmpty(t, login2FA)
	assert.NotEmpty(t, login2FA.AccessToken)
	assert.NotEmpty(t, login2FA.RefreshToken)
}

func login2faNOKOld2FACode(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	user := api.UserSignup{
		Email:        "first.last@example.com",
		FirstName:    "First name",
		LastName:     "Last name",
		Password:     "Som€_$trong_P@$$word",
		TwoFAEnabled: true,
	}

	secret := "JWT_$€CR€T"
	dbConnection := db.Connect(dbConfig)
	defer dbConnection.Close()
	tokenQuery := tokens.New(dbConnection)
	tokenService := tokens.NewService(tokenQuery, []byte(secret), "")
	twoFAQuery := twoFA.New(dbConnection)
	twoFaService := twoFA.NewService("go-jwt-server", twoFAQuery)
	userStorage := New(dbConnection)
	userService := NewService(userStorage, twoFaService, tokenService)

	res, err := userService.Signup(ctx, user)
	assert.NoError(t, err)
	assert.NotEmpty(t, res)
	assert.NotEmpty(t, res.Secret)
	assert.NotEmpty(t, res.QrImage)

	ctx.Header("x-login-source", "test")
	ctx.Header("user-agent", "test")
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.100")
	ctx.Request = req

	userByEmail, err := userStorage.GetEntireUserByEmail(context.Background(), "first.last@example.com")
	assert.NoError(t, err)

	passcode, err := totp.GenerateCode(res.Secret, time.Now().Add(-60*time.Second))
	assert.NoError(t, err)

	login2FA, err := userService.Login2FA(ctx, api.Login2FAParams{}, userByEmail.UserID, passcode)
	assert.Error(t, err)
	assert.Empty(t, login2FA)
}

func login2FANOKUserNotExist(t *testing.T) {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Header("x-login-source", "test")
	ctx.Header("user-agent", "test")
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.100")
	ctx.Request = req

	secret := "JWT_$€CR€T"
	dbConnection := db.Connect(dbConfig)
	defer dbConnection.Close()
	tokenQuery := tokens.New(dbConnection)
	tokenService := tokens.NewService(tokenQuery, []byte(secret), "")
	twoFAQuery := twoFA.New(dbConnection)
	twoFaService := twoFA.NewService("go-jwt-server", twoFAQuery)
	userStorage := New(dbConnection)
	userService := NewService(userStorage, twoFaService, tokenService)

	login2FA, err := userService.Login2FA(ctx, api.Login2FAParams{}, 99999999, "123456")
	assert.Error(t, err)
	assert.Empty(t, login2FA)
}

func TestService_GetUserRolesAndPermissions(t *testing.T) {
	truncateTables()
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Header("x-login-source", "test")
	ctx.Set("User-Email", "first.last@example.com")
	ctx.Header("user-agent", "test")
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.100")
	ctx.Request = req

	request := api.CreateRole{
		Description: "role description",
		Name:        "role_name",
	}
	dbConnection := db.Connect(dbConfig)
	defer dbConnection.Close()
	roleStorage := roles.New(dbConnection)
	roleService := roles.NewService(roleStorage)
	permissionStorage := permissions.New(dbConnection)
	permissionsService := permissions.NewService(permissionStorage)
	roleIds := make([]int64, 10)
	permissionIds := make([]int64, 50)
	for num := range 10 {
		request.Name = fmt.Sprintf("%s_%d", request.Name, num)
		createdRole, err := roleService.CreateNewRole(ctx, api.CreateNewRoleParams{}, "", request)
		assert.NoError(t, err)
		assert.NotEmpty(t, createdRole)
		roleIds[num] = createdRole.Id
		for pn := range 5 {
			permission, err := permissionsService.CreateNewPermission(ctx, api.CreateNewPermissionParams{}, api.CreatePermission{Description: "permission description", Name: fmt.Sprintf("role::create_%d_%d", num, pn)})
			assert.NoError(t, err)
			assert.NotEmpty(t, permission)
			err = roleService.AssignPermissionToRole(ctx, createdRole.Id, api.AssignPermissionToRoleParams{}, api.AssignPermission{Ids: []int64{permission.Id}}, "first.last@example.com")
			assert.NoError(t, err)
			permissionIds[num+pn] = permission.Id
		}
	}
	rolesAndPermissions, err := roleService.ListRolesAndItsPermissions(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, rolesAndPermissions)
	assert.Len(t, rolesAndPermissions, 10)
	for _, rolesAndPermission := range rolesAndPermissions {
		assert.NotEmpty(t, rolesAndPermission)
		assert.Len(t, rolesAndPermission.Roles.Permissions, 5)
	}

	secret := "JWT_$€CR€T"
	tokenQuery := tokens.New(dbConnection)
	tokenService := tokens.NewService(tokenQuery, []byte(secret), "")
	twoFAQuery := twoFA.New(dbConnection)
	twoFaService := twoFA.NewService("go-jwt-server", twoFAQuery)
	userStorage := New(dbConnection)
	userService := NewService(userStorage, twoFaService, tokenService)

	user := api.UserSignup{
		Email:     "first.last@example.com",
		FirstName: "First name",
		LastName:  "Last name",
		Password:  "Som€_$trong_P@$$word",
	}

	res, err := userService.Signup(ctx, user)
	assert.NoError(t, err)
	assert.Empty(t, res)

	userByEmail, err := userStorage.GetEntireUserByEmail(ctx, "first.last@example.com")
	assert.NoError(t, err)
	assert.NotEmpty(t, userByEmail)

	err = userStorage.AssignRolesToUser(ctx, AssignRolesToUserParams{
		UserID:    userByEmail.UserID,
		RoleID:    []int64{roleIds[0], roleIds[1], roleIds[2]},
		CreatedBy: "first.last@example.com",
	})
	assert.NoError(t, err)
	err = userStorage.AssignPermissionToUser(ctx, AssignPermissionToUserParams{
		UserID:       userByEmail.UserID,
		PermissionID: []int64{permissionIds[10], permissionIds[11], permissionIds[12], permissionIds[13]},
		CreatedBy:    "first.last@example.com",
	})
	assert.NoError(t, err)

	userRolesAndPermissions, err := userService.GetUserRolesAndPermissions(ctx, userByEmail.UserID, api.GetRolesOfUserParams{})
	assert.NoError(t, err)
	assert.NotEmpty(t, userRolesAndPermissions)
	assert.Len(t, userRolesAndPermissions.Roles, 3)
	assert.Len(t, userRolesAndPermissions.Permissions, 19)
}

func TestService_AssignRolesToUser(t *testing.T) {
	truncateTables()
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Header("x-login-source", "test")
	ctx.Set("User-Email", "first.last@example.com")
	ctx.Header("user-agent", "test")
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.100")
	ctx.Request = req

	request := api.CreateRole{
		Description: "role description",
		Name:        "role_name",
	}
	dbConnection := db.Connect(dbConfig)
	defer dbConnection.Close()
	roleStorage := roles.New(dbConnection)
	roleService := roles.NewService(roleStorage)
	createdRole, err := roleService.CreateNewRole(ctx, api.CreateNewRoleParams{}, "", request)
	assert.NoError(t, err)
	assert.NotEmpty(t, createdRole)

	rolesAndPermissions, err := roleService.ListRolesAndItsPermissions(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, rolesAndPermissions)
	assert.Len(t, rolesAndPermissions, 1)
	for _, rolesAndPermission := range rolesAndPermissions {
		assert.NotEmpty(t, rolesAndPermission)
		assert.Len(t, rolesAndPermission.Roles.Permissions, 0)
	}

	secret := "JWT_$€CR€T"
	tokenQuery := tokens.New(dbConnection)
	tokenService := tokens.NewService(tokenQuery, []byte(secret), "")
	twoFAQuery := twoFA.New(dbConnection)
	twoFaService := twoFA.NewService("go-jwt-server", twoFAQuery)
	userStorage := New(dbConnection)
	userService := NewService(userStorage, twoFaService, tokenService)

	user := api.UserSignup{
		Email:     "first.last@example.com",
		FirstName: "First name",
		LastName:  "Last name",
		Password:  "Som€_$trong_P@$$word",
	}

	res, err := userService.Signup(ctx, user)
	assert.NoError(t, err)
	assert.Empty(t, res)

	userByEmail, err := userStorage.GetEntireUserByEmail(ctx, "first.last@example.com")
	assert.NoError(t, err)
	assert.NotEmpty(t, userByEmail)

	err = userService.AssignRolesToUser(ctx, userByEmail.UserID, api.AssignRolesToUserParams{}, api.AssignRoleToUser{Roles: []int64{createdRole.Id}}, "first.last@example.com")
	assert.NoError(t, err)
}

func TestService_UnassignRolesToUser(t *testing.T) {
	truncateTables()
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Header("x-login-source", "test")
	ctx.Set("User-Email", "first.last@example.com")
	ctx.Header("user-agent", "test")
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.100")
	ctx.Request = req

	request := api.CreateRole{
		Description: "role description",
		Name:        "role_name",
	}
	dbConnection := db.Connect(dbConfig)
	defer dbConnection.Close()
	roleStorage := roles.New(dbConnection)
	roleService := roles.NewService(roleStorage)
	createdRole, err := roleService.CreateNewRole(ctx, api.CreateNewRoleParams{}, "", request)
	assert.NoError(t, err)
	assert.NotEmpty(t, createdRole)

	rolesAndPermissions, err := roleService.ListRolesAndItsPermissions(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, rolesAndPermissions)
	assert.Len(t, rolesAndPermissions, 1)
	for _, rolesAndPermission := range rolesAndPermissions {
		assert.NotEmpty(t, rolesAndPermission)
		assert.Len(t, rolesAndPermission.Roles.Permissions, 0)
	}

	secret := "JWT_$€CR€T"
	tokenQuery := tokens.New(dbConnection)
	tokenService := tokens.NewService(tokenQuery, []byte(secret), "")
	twoFAQuery := twoFA.New(dbConnection)
	twoFaService := twoFA.NewService("go-jwt-server", twoFAQuery)
	userStorage := New(dbConnection)
	userService := NewService(userStorage, twoFaService, tokenService)

	user := api.UserSignup{
		Email:     "first.last@example.com",
		FirstName: "First name",
		LastName:  "Last name",
		Password:  "Som€_$trong_P@$$word",
	}

	res, err := userService.Signup(ctx, user)
	assert.NoError(t, err)
	assert.Empty(t, res)

	userByEmail, err := userStorage.GetEntireUserByEmail(ctx, "first.last@example.com")
	assert.NoError(t, err)
	assert.NotEmpty(t, userByEmail)

	err = userService.AssignRolesToUser(ctx, userByEmail.UserID, api.AssignRolesToUserParams{}, api.AssignRoleToUser{Roles: []int64{createdRole.Id}}, "first.last@example.com")
	assert.NoError(t, err)

	userRolesAndPermissions, err := userService.GetUserRolesAndPermissions(ctx, userByEmail.UserID, api.GetRolesOfUserParams{})
	assert.NoError(t, err)
	assert.NotEmpty(t, userRolesAndPermissions)
	assert.NotEmpty(t, userRolesAndPermissions.Roles)
	assert.Len(t, userRolesAndPermissions.Roles, 1)

	err = userService.UnassignRolesOfUser(ctx, userByEmail.UserID, createdRole.Id, api.RemoveRolesForUserParams{})
	assert.NoError(t, err)

	userByEmail, err = userStorage.GetEntireUserByEmail(ctx, "first.last@example.com")
	assert.NoError(t, err)
	assert.NotEmpty(t, userByEmail)
	assert.Empty(t, userByEmail.RoleNames)
}
