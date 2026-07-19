package users

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"golang.org/x/crypto/bcrypt"

	"github.com/spdeepak/aegis/server/api"
	"github.com/spdeepak/aegis/server/internal/error"
	"github.com/spdeepak/aegis/server/internal/tokens"
	"github.com/spdeepak/aegis/server/internal/twoFA"
)

type (
	service struct {
		query        Querier
		tokenService tokens.Service
		twoFAService twoFA.Service
	}
	Service interface {
		Signup(ctx *gin.Context, user api.UserSignup) (api.SignUpWith2FAResponse, error)
		Login(ctx *gin.Context, params api.LoginParams, login api.UserLogin) (any, error)
		Login2FA(ctx *gin.Context, params api.Login2FAParams, userId int64, passcode string) (api.LoginSuccessWithJWT, error)
		ChangePassword(ctx *gin.Context, email string, userId int64, changePassword api.ChangePassword) (api.ChangePasswordResponseObject, error)
		RefreshToken(ctx *gin.Context, params api.RefreshParams, refresh api.Refresh) (api.LoginSuccessWithJWT, error)
		GetUserRolesAndPermissions(ctx *gin.Context, id api.Id, params api.GetRolesOfUserParams) (api.UserWithRoles, error)
		AssignRolesToUser(ctx *gin.Context, userId api.Id, params api.AssignRolesToUserParams, assignRoleToUser api.AssignRoleToUser, email string) error
		UnassignRolesOfUser(ctx *gin.Context, userId api.Id, roleId api.RoleId, params api.RemoveRolesForUserParams) error
	}
)

func NewService(query Querier, twoFAService twoFA.Service, tokenService tokens.Service) Service {
	return &service{
		query:        query,
		twoFAService: twoFAService,
		tokenService: tokenService,
	}
}

func (s *service) Signup(ctx *gin.Context, user api.UserSignup) (api.SignUpWith2FAResponse, error) {
	hashedPassword, err := hashPassword(user.Password)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to encrypt password", "error", err)
		return api.SignUpWith2FAResponse{}, err
	}
	email := string(user.Email)
	if !user.TwoFAEnabled {
		userSignup := SignupParams{
			Email:        email,
			FirstName:    user.FirstName,
			LastName:     user.LastName,
			Password:     hashedPassword,
			TwoFaEnabled: user.TwoFAEnabled,
		}
		if err = s.query.Signup(ctx, userSignup); err != nil {
			if err.Error() == "ERROR: duplicate key value violates unique constraint \"users_email_key\" (SQLSTATE 23505)" {
				return api.SignUpWith2FAResponse{}, httperror.New(httperror.UserAlreadyExists)
			}
			return api.SignUpWith2FAResponse{}, httperror.New(httperror.UserSignUpFailed)
		}
		return api.SignUpWith2FAResponse{}, nil
	}

	user2FASetup, err := s.twoFAService.Setup2FA(ctx, email)
	if err != nil {
		return api.SignUpWith2FAResponse{}, err
	}
	userSignupWith2FA := SignupWith2FAParams{
		Secret:       user2FASetup.Secret,
		Url:          user2FASetup.Url,
		Email:        email,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Password:     hashedPassword,
		TwoFaEnabled: user.TwoFAEnabled,
	}
	err = s.query.SignupWith2FA(ctx, userSignupWith2FA)
	if err != nil {
		if err.Error() == "ERROR: duplicate key value violates unique constraint \"users_email_key\" (SQLSTATE 23505)" {
			return api.SignUpWith2FAResponse{}, httperror.New(httperror.UserAlreadyExists)
		}
		return api.SignUpWith2FAResponse{}, httperror.New(httperror.UserSignUpWith2FAFailed)
	}
	return api.SignUpWith2FAResponse{
		QrImage: user2FASetup.QrImage,
		Secret:  user2FASetup.Secret,
	}, nil
}

func (s *service) ChangePassword(ctx *gin.Context, email string, userId int64, changePassword api.ChangePassword) (api.ChangePasswordResponseObject, error) {
	userByEmail, err := s.query.GetEntireUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if !validPassword(changePassword.OldPassword, userByEmail.Password) {
		return api.ChangePassword401Response{}, httperror.New(httperror.InvalidCredentials)
	}
	if userByEmail.TwoFaEnabled {
		if changePassword.TwoFACode == nil {
			slog.ErrorContext(ctx, "Failed to encrypt password", "error", err)
			return api.ChangePassword403Response{}, httperror.New(httperror.TwoFARequired)
		}
		if valid2FA, err := s.twoFAService.Verify2FALogin(ctx, api.Login2FAParams{}, userByEmail.UserID, *changePassword.TwoFACode); err != nil {
			slog.ErrorContext(ctx, "Failed to verify 2fa code", "error", err)
			return nil, err
		} else if !valid2FA {
			slog.ErrorContext(ctx, "Invalid 2fa code")
			return nil, httperror.New(httperror.InvalidTwoFA)
		}
	}
	hashedNewPassword, err := hashPassword(changePassword.NewPassword)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to encrypt password", "error", err)
		return api.ChangePassword400Response{}, fmt.Errorf("hashing password failed")
	}
	changePasswordReq := ChangePasswordParams{
		Password:  hashedNewPassword,
		UserID:    userId,
		UserEmail: email,
	}
	err = s.query.ChangePassword(ctx, changePasswordReq)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to change password", "error", err)
		return nil, httperror.NewWithMetadata(httperror.UserOperationFailed, "Failed to change password")
	}
	return api.ChangePassword200Response{}, nil
}

func (s *service) Login(ctx *gin.Context, params api.LoginParams, login api.UserLogin) (any, error) {
	user, err := s.query.GetEntireUserByEmail(ctx, string(login.Email))
	if err != nil {
		if err.Error() == "no rows in result set" {
			return api.LoginSuccessWithJWT{}, httperror.New(httperror.InvalidCredentials)
		}
		return api.LoginSuccessWithJWT{}, httperror.NewWithMetadata(httperror.UndefinedErrorCode, err.Error())
	}

	if !validPassword(login.Password, user.Password) {
		return api.LoginSuccessWithJWT{}, httperror.New(httperror.InvalidCredentials)
	}
	if user.TwoFaEnabled {
		return s.tokenService.GenerateTempToken(ctx, user.UserID)
	}
	if user.Locked {
		return api.LoginSuccessWithJWT{}, httperror.New(httperror.UserAccountLocked)
	}
	jwtUser := tokens.User{
		ID:        user.UserID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}
	tokenParams := tokens.TokenParams{
		XLoginSource: string(params.XLoginSource),
		UserAgent:    params.UserAgent,
	}
	return s.tokenService.GenerateNewTokenPair(ctx, ctx.ClientIP(), tokenParams, jwtUser, user.RoleNames, user.PermissionNames)
}

func (s *service) Login2FA(ctx *gin.Context, params api.Login2FAParams, userId int64, passcode string) (api.LoginSuccessWithJWT, error) {
	isValid, err := s.twoFAService.Verify2FALogin(ctx, params, userId, passcode)
	if err != nil {
		return api.LoginSuccessWithJWT{}, httperror.NewWithMetadata(httperror.InvalidTwoFA, err.Error())
	}
	if !isValid {
		return api.LoginSuccessWithJWT{}, httperror.New(httperror.InvalidTwoFA)
	}

	user, err := s.query.GetUserById(ctx, userId)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return api.LoginSuccessWithJWT{}, httperror.New(httperror.InvalidCredentials)
		}
		return api.LoginSuccessWithJWT{}, httperror.NewWithStatus(httperror.UndefinedErrorCode, err.Error(), http.StatusBadRequest)
	}

	if user.Locked {
		return api.LoginSuccessWithJWT{}, httperror.New(httperror.UserAccountLocked)
	}

	jwtUser := tokens.User{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}
	tokenParams := tokens.TokenParams{
		XLoginSource: string(params.XLoginSource),
		UserAgent:    params.UserAgent,
	}
	return s.tokenService.GenerateNewTokenPair(ctx, ctx.ClientIP(), tokenParams, jwtUser, nil, nil)
}

func (s *service) RefreshToken(ctx *gin.Context, params api.RefreshParams, refresh api.Refresh) (api.LoginSuccessWithJWT, error) {
	claims, err := s.tokenService.ValidateRefreshToken(ctx, ctx.ClientIP(), params, refresh.RefreshToken)
	if err != nil {
		return api.LoginSuccessWithJWT{}, err
	}
	email, ok := claims["email"].(string)
	if !ok {
		return api.LoginSuccessWithJWT{}, httperror.NewWithMetadata(httperror.InvalidRefreshToken, "Invalid token claims")
	}

	user, err := s.query.GetEntireUserByEmail(ctx, email)
	if err != nil {
		return api.LoginSuccessWithJWT{}, httperror.NewWithMetadata(httperror.InvalidRefreshToken, "Invalid token claims")
	}
	jwtUser := tokens.User{
		ID:        user.UserID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}

	tokenParams := tokens.TokenParams{
		XLoginSource: string(params.XLoginSource),
		UserAgent:    params.UserAgent,
	}
	return s.tokenService.RefreshAndInvalidateToken(ctx, ctx.ClientIP(), tokenParams, refresh, jwtUser, user.RoleNames, user.PermissionNames)
}

func (s *service) GetUserRolesAndPermissions(ctx *gin.Context, id api.Id, params api.GetRolesOfUserParams) (api.UserWithRoles, error) {
	userRolesAndPermissions, err := s.query.GetUserRolesAndPermissionsFromID(ctx, id)
	if err != nil {
		return api.UserWithRoles{}, err
	}
	return api.UserWithRoles{
		Id:          userRolesAndPermissions.UserID,
		Email:       openapi_types.Email(userRolesAndPermissions.Email),
		FirstName:   userRolesAndPermissions.FirstName,
		LastName:    userRolesAndPermissions.LastName,
		Permissions: userRolesAndPermissions.PermissionNames,
		Roles:       userRolesAndPermissions.RoleNames,
	}, nil
}

func (s *service) AssignRolesToUser(ctx *gin.Context, userId api.Id, params api.AssignRolesToUserParams, assignRoleToUser api.AssignRoleToUser, email string) error {
	rolesIds := make([]int64, len(assignRoleToUser.Roles))
	for index, id := range assignRoleToUser.Roles {
		rolesIds[index] = id
	}
	assignRolesToUser := AssignRolesToUserParams{
		UserID:    userId,
		RoleID:    rolesIds,
		CreatedBy: email,
	}
	if err := s.query.AssignRolesToUser(ctx, assignRolesToUser); err != nil {
		var pgerr *pgconn.PgError
		if errors.As(err, &pgerr) {
			if pgerr.Code == "23503" {
				return httperror.New(httperror.RoleDoesntExist)
			}
		}
		return err
	}
	return nil
}

func (s *service) UnassignRolesOfUser(ctx *gin.Context, userId api.Id, roleId api.RoleId, params api.RemoveRolesForUserParams) error {
	unassignRolesToUser := UnassignRolesToUserParams{
		UserID: userId,
		RoleID: roleId,
	}
	err := s.query.UnassignRolesToUser(ctx, unassignRolesToUser)
	if err != nil {
		return err
	}
	return nil
}

// hashPassword hashes the plaintext password using bcrypt
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// validPassword compares plaintext and hashed password
func validPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
