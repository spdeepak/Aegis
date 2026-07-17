package roles

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/spdeepak/aegis/server/api"
	"github.com/spdeepak/aegis/server/internal/error"
)

type service struct {
	storage Querier
}

type Service interface {
	CreateNewRole(ctx context.Context, params api.CreateNewRoleParams, email string, request api.CreateRole) (api.RoleResponse, error)
	DeleteRoleById(ctx context.Context, id int64) error
	GetRoleById(ctx context.Context, id int64) (api.RoleResponse, error)
	ListRoles(ctx context.Context) ([]api.RoleResponse, error)
	UpdateRoleById(ctx context.Context, id api.Id, email string, params api.UpdateRoleByIdParams, req api.UpdateRole) (api.RoleResponse, error)
	AssignPermissionToRole(ctx context.Context, roleId api.Id, params api.AssignPermissionToRoleParams, assignPermission api.AssignPermission, email string) error
	UnassignPermissionFromRole(ctx context.Context, roleId api.RoleId, permissionId api.PermissionId) error
	ListRolesAndItsPermissions(ctx context.Context) ([]api.RolesAndPermissionResponse, error)
}

func NewService(storage Querier) Service {
	return &service{
		storage: storage,
	}
}

func (s *service) CreateNewRole(ctx context.Context, params api.CreateNewRoleParams, email string, request api.CreateRole) (api.RoleResponse, error) {
	createNewRole := CreateNewRoleParams{
		Name:        request.Name,
		Description: request.Description,
		CreatedBy:   email,
	}
	createdNewRole, err := s.storage.CreateNewRole(ctx, createNewRole)
	if err != nil {
		if err.Error() == "ERROR: duplicate key value violates unique constraint \"roles_name_key\" (SQLSTATE 23505)" {
			return api.RoleResponse{}, httperror.New(httperror.RoleAlreadyExists)
		}
		return api.RoleResponse{}, httperror.NewWithMetadata(httperror.RoleCreationFailed, err.Error())
	}
	return api.RoleResponse{
		CreatedAt:   createdNewRole.CreatedAt,
		CreatedBy:   createdNewRole.CreatedBy,
		Description: createdNewRole.Description,
		Id:          createdNewRole.ID,
		Name:        createdNewRole.Name,
		UpdatedAt:   createdNewRole.UpdatedAt,
		UpdatedBy:   createdNewRole.UpdatedBy,
	}, nil
}

func (s *service) DeleteRoleById(ctx context.Context, id int64) error {
	return s.storage.DeleteRoleById(ctx, id)
}

func (s *service) GetRoleById(ctx context.Context, id int64) (api.RoleResponse, error) {
	roleById, err := s.storage.GetRoleById(ctx, id)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return api.RoleResponse{}, httperror.New(httperror.RoleDoesntExist)
		}
		return api.RoleResponse{}, httperror.NewWithDescription("Couldn't fetch Role for given ID", http.StatusInternalServerError)
	}
	return api.RoleResponse{
		CreatedAt:   roleById.CreatedAt,
		CreatedBy:   roleById.CreatedBy,
		Description: roleById.Description,
		Name:        roleById.Name,
		UpdatedAt:   roleById.UpdatedAt,
		UpdatedBy:   roleById.UpdatedBy,
	}, nil
}

func (s *service) ListRoles(ctx context.Context) ([]api.RoleResponse, error) {
	allRoles, err := s.storage.ListRoles(ctx)
	if err != nil {
		return nil, err
	}
	roles := make([]api.RoleResponse, len(allRoles))
	for index, role := range allRoles {
		roles[index] = api.RoleResponse{
			CreatedAt:   role.CreatedAt,
			CreatedBy:   role.CreatedBy,
			Description: role.Description,
			Id:          role.ID,
			Name:        role.Name,
			UpdatedAt:   role.UpdatedAt,
			UpdatedBy:   role.UpdatedBy,
		}
	}
	return roles, nil
}

func (s *service) UpdateRoleById(ctx context.Context, id api.Id, email string, params api.UpdateRoleByIdParams, req api.UpdateRole) (api.RoleResponse, error) {
	updateRoleByIdParam := UpdateRoleByIdParams{
		ID:        id,
		UpdatedBy: email,
	}
	if req.Description != nil && *req.Description != "" {
		updateRoleByIdParam.Description = pgtype.Text{
			String: *req.Description,
			Valid:  true,
		}
	}
	if req.Name != nil && *req.Name != "" {
		updateRoleByIdParam.Name = pgtype.Text{
			String: *req.Name,
			Valid:  true,
		}
	}
	updatedRole, err := s.storage.UpdateRoleById(ctx, updateRoleByIdParam)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return api.RoleResponse{}, httperror.New(httperror.RoleDoesntExist)
		}
		return api.RoleResponse{}, httperror.NewWithDescription("Couldn't fetch Role for given ID", http.StatusInternalServerError)
	}
	return api.RoleResponse{
		CreatedAt:   updatedRole.CreatedAt,
		CreatedBy:   updatedRole.CreatedBy,
		Description: updatedRole.Description,
		Name:        updatedRole.Name,
		UpdatedAt:   updatedRole.UpdatedAt,
		UpdatedBy:   updatedRole.UpdatedBy,
	}, nil
}

func (s *service) AssignPermissionToRole(ctx context.Context, roleId api.Id, params api.AssignPermissionToRoleParams, assignPermission api.AssignPermission, email string) error {
	permissionIds := make([]int64, len(assignPermission.Ids))
	for index, id := range assignPermission.Ids {
		permissionIds[index] = id
	}
	assignPermissionsToRole := AssignPermissionsParams{
		RoleID:       roleId,
		PermissionID: permissionIds,
		CreatedBy:    email,
	}
	if err := s.storage.AssignPermissions(ctx, assignPermissionsToRole); err != nil {
		return httperror.NewWithDescription("Failed to assign permission to role", http.StatusInternalServerError)
	}
	return nil
}

func (s *service) UnassignPermissionFromRole(ctx context.Context, roleId api.RoleId, permissionId api.PermissionId) error {
	UnassignPermissionFromRole := UnAssignPermissionParams{
		RoleID:       roleId,
		PermissionID: permissionId,
	}
	if err := s.storage.UnAssignPermission(ctx, UnassignPermissionFromRole); err != nil {
		return httperror.NewWithDescription("Failed to assign permission to role", http.StatusInternalServerError)
	}
	return nil
}

func (s *service) ListRolesAndItsPermissions(ctx context.Context) ([]api.RolesAndPermissionResponse, error) {
	rolesAndItsPermissions, err := s.storage.ListRolesAndItsPermissions(ctx)
	if err != nil {
		return nil, err
	}

	roleIdRolePermissionMap := make(map[int64]*api.RolesAndPermissionResponse)

	for _, rolePermission := range rolesAndItsPermissions {
		rolePermissionResponse, exists := roleIdRolePermissionMap[rolePermission.RoleID]
		if !exists {
			var rolePerm api.RolesAndPermissionResponse
			rolePerm.Roles.CreatedAt = rolePermission.RoleCreatedAt
			rolePerm.Roles.CreatedBy = rolePermission.RoleCreatedBy
			rolePerm.Roles.Description = rolePermission.RoleDescription
			rolePerm.Roles.Id = rolePermission.RoleID
			rolePerm.Roles.Name = rolePermission.RoleName
			rolePerm.Roles.UpdatedAt = rolePermission.RoleUpdatedAt
			rolePerm.Roles.UpdatedBy = rolePermission.RoleUpdatedBy
			roleIdRolePermissionMap[rolePermission.RoleID] = &rolePerm
			rolePermissionResponse = roleIdRolePermissionMap[rolePermission.RoleID]
		}

		if rolePermission.PermissionID.Valid {
			rolePermissionResponse.Roles.Permissions = append(rolePermissionResponse.Roles.Permissions, api.PermissionResponse{
				CreatedAt:   rolePermission.PermissionCreatedAt.Time,
				CreatedBy:   rolePermission.PermissionCreatedBy.String,
				Description: rolePermission.PermissionDescription.String,
				Id:          rolePermission.PermissionID.Int64,
				Name:        rolePermission.PermissionName.String,
				UpdatedAt:   rolePermission.PermissionUpdatedAt.Time,
				UpdatedBy:   rolePermission.PermissionUpdatedBy.String,
			})
		}
	}

	rolePermissionResponse := make([]api.RolesAndPermissionResponse, 0)
	for _, role := range roleIdRolePermissionMap {
		rolePermissionResponse = append(rolePermissionResponse, *role)
	}

	return rolePermissionResponse, nil
}
