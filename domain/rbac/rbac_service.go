package rbac

import (
	"better-console-backend/dtos"
	"context"
)

type RoleBasedAccessControlService struct {
}

func (RoleBasedAccessControlService) CreatePermission(ctx context.Context, permissionInformation dtos.PermissionInformation) error {
	permissionEntity := NewPermissionEntity(permissionInformation)
	return permissionRepository{}.Create(ctx, &permissionEntity)
}

func (RoleBasedAccessControlService) GetPermissions(ctx context.Context, filters map[string]interface{}, pageable dtos.Pageable) ([]PermissionEntity, int64, error) {
	return permissionRepository{}.FindAll(ctx, filters, pageable)
}

func (RoleBasedAccessControlService) CreateRole(ctx context.Context, roleInformation dtos.RoleInformation) error {
	roleEntity, err := NewRoleEntity(ctx, roleInformation)
	if err != nil {
		return err
	}

	return roleRepository{}.Create(ctx, &roleEntity)
}

func (RoleBasedAccessControlService) GetRoles(ctx context.Context, filters map[string]interface{}, pageable dtos.Pageable) ([]RoleEntity, int64, error) {
	return roleRepository{}.FindAll(ctx, filters, pageable)
}

func (RoleBasedAccessControlService) UpdatePermission(ctx context.Context, permissionId uint, permissionInformation dtos.PermissionInformation) error {
	repository := permissionRepository{}
	permissionEntity, err := repository.FindById(ctx, permissionId)
	if err != nil {
		return err
	}

	if err := permissionEntity.Update(ctx, permissionInformation); err != nil {
		return err
	}

	return repository.Save(ctx, permissionEntity)
}

func (RoleBasedAccessControlService) DeletePermission(ctx context.Context, permissionId uint) error {
	repository := permissionRepository{}

	permissionEntity, err := repository.FindById(ctx, permissionId)
	if err != nil {
		return err
	}

	if err := permissionEntity.Deletable(); err != nil {
		return err
	}

	return repository.Delete(ctx, permissionEntity)
}

func (RoleBasedAccessControlService) DeleteRole(ctx context.Context, roleId uint) error {
	repository := roleRepository{}

	roleEntity, err := repository.FindById(ctx, roleId)
	if err != nil {
		return err
	}

	if err := roleEntity.Deletable(); err != nil {
		return err
	}

	return repository.Delete(ctx, roleEntity)
}

func (RoleBasedAccessControlService) UpdateRole(ctx context.Context, roleId uint, roleInformation dtos.RoleInformation) error {
	repository := roleRepository{}
	roleEntity, err := repository.FindById(ctx, roleId)
	if err != nil {
		return err
	}

	if err := roleEntity.Update(ctx, roleInformation); err != nil {
		return err
	}

	return repository.Save(ctx, &roleEntity)
}
