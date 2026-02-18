package permissions

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/spdeepak/go-jwt-server/api"
	"github.com/spdeepak/go-jwt-server/internal/error"
)

const emailHeader = "User-Email"

type service struct {
	storage Querier
}

type Service interface {
	CreateNewPermission(ctx *gin.Context, params api.CreateNewPermissionParams, request api.CreatePermission) (api.PermissionResponse, error)
	DeletePermissionById(ctx *gin.Context, id int64) error
	GetPermissionById(ctx *gin.Context, id int64) (api.PermissionResponse, error)
	ListPermissions(ctx *gin.Context) ([]api.PermissionResponse, error)
	UpdatePermissionById(ctx *gin.Context, id api.Id, params api.UpdatePermissionByIdParams, req api.UpdatePermission) (api.PermissionResponse, error)
}

func NewService(storage Querier) Service {
	return &service{
		storage: storage,
	}
}

func (s *service) CreateNewPermission(ctx *gin.Context, params api.CreateNewPermissionParams, request api.CreatePermission) (api.PermissionResponse, error) {
	email, _ := ctx.Get(emailHeader)
	createNewPermissionParam := CreateNewPermissionParams{
		Name:        request.Name,
		Description: request.Description,
		CreatedBy:   email.(string),
	}
	createdNewPermission, err := s.storage.CreateNewPermission(ctx, createNewPermissionParam)
	if err != nil {
		if err.Error() == "ERROR: duplicate key value violates unique constraint \"permissions_name_key\" (SQLSTATE 23505)" {
			return api.PermissionResponse{}, httperror.New(httperror.PermissionAlreadyExists)
		}
		return api.PermissionResponse{}, httperror.NewWithMetadata(httperror.PermissionCreationFailed, err.Error())
	}
	return api.PermissionResponse{
		CreatedAt:   createdNewPermission.CreatedAt,
		CreatedBy:   createdNewPermission.CreatedBy,
		Description: createdNewPermission.Description,
		Id:          createdNewPermission.ID,
		Name:        createdNewPermission.Name,
		UpdatedAt:   createdNewPermission.UpdatedAt,
		UpdatedBy:   createdNewPermission.UpdatedBy,
	}, nil
}

func (s *service) DeletePermissionById(ctx *gin.Context, id int64) error {
	return s.storage.DeletePermissionById(ctx, id)
}

func (s *service) GetPermissionById(ctx *gin.Context, id int64) (api.PermissionResponse, error) {
	permissionById, err := s.storage.GetPermissionById(ctx, id)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return api.PermissionResponse{}, httperror.New(httperror.PermissionDoesntExist)
		}
		return api.PermissionResponse{}, httperror.NewWithDescription("Couldn't fetch Permission for given ID", http.StatusInternalServerError)
	}
	return api.PermissionResponse{
		CreatedAt:   permissionById.CreatedAt,
		CreatedBy:   permissionById.CreatedBy,
		Description: permissionById.Description,
		Name:        permissionById.Name,
		UpdatedAt:   permissionById.UpdatedAt,
		UpdatedBy:   permissionById.UpdatedBy,
	}, nil
}

func (s *service) ListPermissions(ctx *gin.Context) ([]api.PermissionResponse, error) {
	listedPermissions, err := s.storage.ListPermissions(ctx)
	if err != nil {
		return nil, err
	}
	permissions := make([]api.PermissionResponse, len(listedPermissions))
	for index, permission := range listedPermissions {
		permissions[index] = api.PermissionResponse{
			CreatedAt:   permission.CreatedAt,
			CreatedBy:   permission.CreatedBy,
			Description: permission.Description,
			Id:          permission.ID,
			Name:        permission.Name,
			UpdatedAt:   permission.UpdatedAt,
			UpdatedBy:   permission.UpdatedBy,
		}
	}
	return permissions, nil
}

func (s *service) UpdatePermissionById(ctx *gin.Context, id api.Id, params api.UpdatePermissionByIdParams, req api.UpdatePermission) (api.PermissionResponse, error) {
	email, _ := ctx.Get(emailHeader)
	updatePermissionByIdParam := UpdatePermissionByIdParams{
		ID:        id,
		UpdatedBy: email.(string),
	}
	if req.Description != nil && *req.Description != "" {
		updatePermissionByIdParam.Description = pgtype.Text{
			String: *req.Description,
			Valid:  true,
		}
	}
	if req.Name != nil && *req.Name != "" {
		updatePermissionByIdParam.Name = pgtype.Text{
			String: *req.Name,
			Valid:  true,
		}
	}
	updatedPermission, err := s.storage.UpdatePermissionById(ctx, updatePermissionByIdParam)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return api.PermissionResponse{}, httperror.New(httperror.PermissionDoesntExist)
		}
		return api.PermissionResponse{}, httperror.NewWithDescription("Couldn't fetch Permission for given ID", http.StatusInternalServerError)
	}
	return api.PermissionResponse{
		CreatedAt:   updatedPermission.CreatedAt,
		CreatedBy:   updatedPermission.CreatedBy,
		Description: updatedPermission.Description,
		Name:        updatedPermission.Name,
		UpdatedAt:   updatedPermission.UpdatedAt,
		UpdatedBy:   updatedPermission.UpdatedBy,
	}, nil
}
