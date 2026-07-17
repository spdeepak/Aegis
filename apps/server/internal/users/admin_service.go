package users

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/spdeepak/aegis/server/api"
	httperror "github.com/spdeepak/aegis/server/internal/error"
)

type (
	adminService struct {
		storage Querier
	}
	AdminService interface {
		GetListOfUsers(ctx context.Context, params api.GetListOfUsersParams) ([]api.UserDetails, error)
		LockUserById(ctx context.Context, id int64, params api.LockUserParams) error
		UnlockUserById(ctx context.Context, id int64, params api.UnlockUserParams) error
		DisableUserById(ctx context.Context, id int64, params api.DisableUserParams) error
		EnableUserById(ctx context.Context, id int64, params api.EnableUserParams) error
	}
)

func NewAdminService(storage Querier) AdminService {
	return &adminService{
		storage: storage,
	}
}

func (a *adminService) GetListOfUsers(ctx context.Context, params api.GetListOfUsersParams) ([]api.UserDetails, error) {
	repoParams := SearchAndGetUserDetailsParams{}
	if params.FirstName != nil {
		repoParams.FirstName = pgtype.Text{
			String: *params.FirstName,
			Valid:  true,
		}
	} else {
		repoParams.FirstName = pgtype.Text{
			Valid: false,
		}
	}
	if params.LastName != nil {
		repoParams.LastName = pgtype.Text{
			String: *params.LastName,
			Valid:  true,
		}
	} else {
		repoParams.LastName = pgtype.Text{
			Valid: false,
		}
	}
	if params.Email != nil {
		repoParams.Email = pgtype.Text{
			String: string(*params.Email),
			Valid:  true,
		}
	} else {
		repoParams.Email = pgtype.Text{
			Valid: false,
		}
	}
	if params.Page != nil {
		repoParams.Page = int32(*params.Page)
	}
	if params.Size != nil {
		repoParams.Size = int32(*params.Size)
	}
	details, err := a.storage.SearchAndGetUserDetails(ctx, repoParams)
	if err != nil {
		return nil, err
	}
	userDetails := make([]api.UserDetails, len(details))
	for index, detail := range details {
		userDetails[index] = api.UserDetails{
			Email:       openapi_types.Email(detail.Email),
			FirstName:   detail.FirstName,
			Id:          detail.UserID,
			LastName:    detail.LastName,
			Permissions: detail.Permissions,
			Roles:       detail.Roles,
		}
	}
	return userDetails, nil
}

func (a *adminService) LockUserById(ctx context.Context, id int64, params api.LockUserParams) error {
	_, err := a.storage.LockUserById(ctx, LockUserByIdParams{
		UserID:    id,
		ActorID:   ctx.Value("User-ID").(int64),
		IpAddress: ctx.Value("user-ip").(string),
		UserAgent: params.UserAgent,
	})
	if err != nil && err.Error() == "no rows in result set" {
		return httperror.NewWithMetadata(httperror.UserNotFound, "Invalid user id")
	} else if err != nil {
		return httperror.NewWithMetadata(httperror.UserOperationFailed, "Failed to lock user")
	}
	return nil
}

func (a *adminService) UnlockUserById(ctx context.Context, id int64, params api.UnlockUserParams) error {
	_, err := a.storage.UnlockUserById(ctx, UnlockUserByIdParams{
		UserID:    id,
		ActorID:   ctx.Value("User-ID").(int64),
		IpAddress: ctx.Value("user-ip").(string),
		UserAgent: params.UserAgent,
	})
	if err != nil && err.Error() == "no rows in result set" {
		return httperror.NewWithMetadata(httperror.UserNotFound, "Invalid user id")
	} else if err != nil {
		return httperror.NewWithMetadata(httperror.UserOperationFailed, "Failed to unlock user")
	}
	return nil
}

func (a *adminService) DisableUserById(ctx context.Context, id int64, params api.DisableUserParams) error {
	_, err := a.storage.DisableUserById(ctx, DisableUserByIdParams{
		UserID:    id,
		ActorID:   ctx.Value("User-ID").(int64),
		IpAddress: ctx.Value("user-ip").(string),
		UserAgent: params.UserAgent,
	})
	if err != nil && err.Error() == "no rows in result set" {
		return httperror.NewWithMetadata(httperror.UserNotFound, "Invalid user id")
	} else if err != nil {
		return httperror.NewWithMetadata(httperror.UserOperationFailed, "Failed to disable user")
	}
	return nil
}

func (a *adminService) EnableUserById(ctx context.Context, id int64, params api.EnableUserParams) error {
	_, err := a.storage.EnableUserById(ctx, EnableUserByIdParams{
		UserID:    id,
		ActorID:   ctx.Value("User-ID").(int64),
		IpAddress: ctx.Value("user-ip").(string),
		UserAgent: params.UserAgent,
	})
	if err != nil && err.Error() == "no rows in result set" {
		return httperror.NewWithMetadata(httperror.UserNotFound, "Invalid user id")
	} else if err != nil {
		return httperror.NewWithMetadata(httperror.UserOperationFailed, "Failed to enable user")
	}
	return nil
}
