package roles

import (
	"context"
	"fmt"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/spdeepak/aegis/api"
	"github.com/spdeepak/aegis/internal/config"
	"github.com/spdeepak/aegis/internal/db"
	"github.com/spdeepak/aegis/internal/permissions"
	permissionsRepo "github.com/spdeepak/aegis/internal/permissions"
)

var roleStorage Querier
var permissionStorage permissionsRepo.Querier
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
	MaxOpenConns:      1,
	MaxIdleConns:      1,
	ConnMaxLifetime:   10 * time.Minute,
	ConnMaxIdleTime:   2 * time.Minute,
	HealthCheckPeriod: 1 * time.Minute,
}

func TestMain(m *testing.M) {
	dbConnection := db.Connect(dbConfig)
	roleStorage = New(dbConnection)
	permissionStorage = permissionsRepo.New(dbConnection)
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
	assert.NoError(t, err)
}

func TestService_CreateNewRole(t *testing.T) {
	truncateTables()
	t.Run("Create New Role OK", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Set("User-Email", "first.last@example.com")
		request := api.CreateRole{
			Description: "role description",
			Name:        "role_name",
		}
		roleService := NewService(roleStorage)
		createdRole, err := roleService.CreateNewRole(ctx, api.CreateNewRoleParams{}, "first.last@example.com", request)
		assert.NoError(t, err)
		assert.NotEmpty(t, createdRole)
	})
	t.Run("Create New Role NOK duplicate", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Set("User-Email", "first.last@example.com")
		request := api.CreateRole{
			Description: "role description",
			Name:        "role_name",
		}
		roleService := NewService(roleStorage)
		createdRole, err := roleService.CreateNewRole(ctx, api.CreateNewRoleParams{}, "first.last@example.com", request)
		assert.Error(t, err)
		assert.Empty(t, createdRole)
	})
}

func TestService_DeleteRole(t *testing.T) {
	truncateTables()
	t.Run("Delete Role OK", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Set("User-Email", "first.last@example.com")
		request := api.CreateRole{
			Description: "role description",
			Name:        "role_name",
		}
		roleService := NewService(roleStorage)
		createdRole, err := roleService.CreateNewRole(ctx, api.CreateNewRoleParams{}, "first.last@example.com", request)
		assert.NoError(t, err)
		assert.NotEmpty(t, createdRole)
		err = roleService.DeleteRoleById(ctx, createdRole.Id)
		assert.NoError(t, err)
	})
	t.Run("Delete Role OK not exists", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Set("User-Email", "first.last@example.com")
		roleService := NewService(roleStorage)
		err := roleService.DeleteRoleById(ctx, 999999999)
		assert.NoError(t, err)
	})
}

func TestService_ListRoles(t *testing.T) {
	truncateTables()
	t.Run("List Roles OK", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Set("User-Email", "first.last@example.com")
		request := api.CreateRole{
			Description: "role description",
			Name:        "role_name",
		}
		roleService := NewService(roleStorage)
		createdRole, err := roleService.CreateNewRole(ctx, api.CreateNewRoleParams{}, "first.last@example.com", request)
		assert.NoError(t, err)
		assert.NotEmpty(t, createdRole)
		roles, err := roleService.ListRoles(ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, roles)
		assert.Equal(t, request.Name, roles[0].Name)
		assert.Equal(t, request.Description, roles[0].Description)
	})
	truncateTables()
	t.Run("List Roles Empty", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Set("User-Email", "first.last@example.com")
		roleService := NewService(roleStorage)
		createdRole, err := roleService.ListRoles(ctx)
		assert.NoError(t, err)
		assert.Empty(t, createdRole)
	})
}

func TestService_GetRoleById(t *testing.T) {
	truncateTables()
	t.Run("Get Role by ID OK", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Set("User-Email", "first.last@example.com")
		request := api.CreateRole{
			Description: "role description",
			Name:        "role_name",
		}
		roleService := NewService(roleStorage)
		createdRole, err := roleService.CreateNewRole(ctx, api.CreateNewRoleParams{}, "first.last@example.com", request)
		assert.NoError(t, err)
		assert.NotEmpty(t, createdRole)
		role, err := roleService.GetRoleById(ctx, createdRole.Id)
		assert.NoError(t, err)
		assert.NotEmpty(t, role)
		assert.Equal(t, request.Name, role.Name)
		assert.Equal(t, request.Description, role.Description)
	})
	truncateTables()
	t.Run("Get Role by ID NOK", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Set("User-Email", "first.last@example.com")
		roleService := NewService(roleStorage)
		role, err := roleService.GetRoleById(ctx, 999999999)
		assert.Error(t, err)
		assert.Empty(t, role)
	})
}

func TestService_UpdateRoleById(t *testing.T) {
	truncateTables()
	t.Run("Get Role by ID OK", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Set("User-Email", "first.last@example.com")
		request := api.CreateRole{
			Description: "role description",
			Name:        "role_name",
		}
		roleService := NewService(roleStorage)
		createdRole, err := roleService.CreateNewRole(ctx, api.CreateNewRoleParams{}, "first.last@example.com", request)
		assert.NoError(t, err)
		assert.NotEmpty(t, createdRole)
		updatedRoleDescription := "changed role description"
		updatedName := "updated_role_name"
		role, err := roleService.UpdateRoleById(ctx, createdRole.Id, "first.last@example.com", api.UpdateRoleByIdParams{}, api.UpdateRole{Description: &updatedRoleDescription, Name: &updatedName})
		assert.NoError(t, err)
		assert.NotEmpty(t, role)
		assert.Equal(t, updatedRoleDescription, role.Description)
		assert.Equal(t, updatedName, role.Name)
		assert.Equal(t, role.CreatedAt, createdRole.CreatedAt)
		assert.NotEqual(t, role.UpdatedAt, createdRole.UpdatedAt)
	})
	truncateTables()
	t.Run("Get Role by ID NOK", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Set("User-Email", "first.last@example.com")
		roleService := NewService(roleStorage)

		updatedRoleDescription := "changed role description"
		updatedName := "updated_role_name"
		role, err := roleService.UpdateRoleById(ctx, 999999999, "first.last@example.com", api.UpdateRoleByIdParams{}, api.UpdateRole{Description: &updatedRoleDescription, Name: &updatedName})
		assert.Error(t, err)
		assert.Empty(t, role)
	})
}

func TestService_AssignPermissionToRole(t *testing.T) {
	truncateTables()
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Set("User-Email", "first.last@example.com")
	request := api.CreateRole{
		Description: "role description",
		Name:        "role_name",
	}
	roleService := NewService(roleStorage)
	createdRole, err := roleService.CreateNewRole(ctx, api.CreateNewRoleParams{}, "first.last@example.com", request)
	assert.NoError(t, err)
	assert.NotEmpty(t, createdRole)
	permissionsService := permissions.NewService(permissionStorage)
	permission, err := permissionsService.CreateNewPermission(ctx, api.CreateNewPermissionParams{}, api.CreatePermission{Description: "permission description", Name: "role::create"})
	assert.NoError(t, err)
	assert.NotEmpty(t, permission)

	err = roleService.AssignPermissionToRole(ctx, createdRole.Id, api.AssignPermissionToRoleParams{}, api.AssignPermission{Ids: []int64{permission.Id}}, "first.last@example.com")
	assert.NoError(t, err)
}

func TestService_UnassignPermissionToRole(t *testing.T) {
	truncateTables()
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Set("User-Email", "first.last@example.com")
	request := api.CreateRole{
		Description: "role description",
		Name:        "role_name",
	}
	roleService := NewService(roleStorage)
	createdRole, err := roleService.CreateNewRole(ctx, api.CreateNewRoleParams{}, "first.last@example.com", request)
	assert.NoError(t, err)
	assert.NotEmpty(t, createdRole)
	permissionsService := permissions.NewService(permissionStorage)
	permission, err := permissionsService.CreateNewPermission(ctx, api.CreateNewPermissionParams{}, api.CreatePermission{Description: "permission description", Name: "role::create"})
	assert.NoError(t, err)
	assert.NotEmpty(t, permission)

	err = roleService.AssignPermissionToRole(ctx, createdRole.Id, api.AssignPermissionToRoleParams{}, api.AssignPermission{Ids: []int64{permission.Id}}, "first.last@example.com")
	assert.NoError(t, err)
	err = roleService.UnassignPermissionFromRole(ctx, createdRole.Id, permission.Id)
	assert.NoError(t, err)
}

func TestService_ListRolesAndItsPermissions(t *testing.T) {
	truncateTables()
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Set("User-Email", "first.last@example.com")
	request := api.CreateRole{
		Description: "role description",
		Name:        "role_name",
	}
	roleService := NewService(roleStorage)
	permissionsService := permissions.NewService(permissionStorage)

	for num := range 10 {
		request.Name = fmt.Sprintf("%s_%d", request.Name, num)
		createdRole, err := roleService.CreateNewRole(ctx, api.CreateNewRoleParams{}, "first.last@example.com", request)
		assert.NoError(t, err)
		assert.NotEmpty(t, createdRole)
		for pn := range 5 {
			permission, err := permissionsService.CreateNewPermission(ctx, api.CreateNewPermissionParams{}, api.CreatePermission{Description: "permission description", Name: fmt.Sprintf("role::create_%d_%d", num, pn)})
			assert.NoError(t, err)
			assert.NotEmpty(t, permission)
			err = roleService.AssignPermissionToRole(ctx, createdRole.Id, api.AssignPermissionToRoleParams{}, api.AssignPermission{Ids: []int64{permission.Id}}, "first.last@example.com")
			assert.NoError(t, err)
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
}
