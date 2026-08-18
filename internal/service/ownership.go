package service

import (
	"errors"

	"drone-platform/internal/domain"
)

// ErrNotOwner 资源属主校验失败（用户侧写接口越权防护）。
var ErrNotOwner = errors.New("仅资源属主可执行此操作")

// canMutate 属主校验：本人 或 平台/协会管理员 可改删资源。
func canMutate(a domain.Actor, ownerID string) bool {
	return a.ID == ownerID ||
		a.Role == domain.RolePlatformAdmin ||
		a.Role == domain.RoleAssociationAdmin
}
