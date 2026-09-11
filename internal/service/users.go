package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/crypto/bcrypt"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// UserPurgeRetention 注销后的缓冲期：账号立即失效，账号行保留 7 天后由后台任务物理清除。
//
// 产品口径（2026-09-11 定稿）：注销**不可恢复**，只保留一个缓冲期，避免数据瞬间不可追。
// 因此后台只有一个动作"删除"，不再区分"注销 / 彻底删除"，也不提供恢复入口。
const UserPurgeRetention = 7 * 24 * time.Hour

// UserService 后台用户账号处置。
//
// 策略依据（生产库实测）：只有 refresh_tokens / user_roles / work_orders 四个外键指向 users；
// 另有 46 张业务表用文本列记录用户 ID 且**没有外键**（demands.publisher_id、posts.author_id…），
// 所以账号行删除不会连带删除内容，只会让内容失去作者（孤儿数据，生产现存 12 需求 + 3 动态）。
// 内容清理不在本服务范围（要清得单独定策略）。
type UserService struct {
	users        repository.UserRepository
	files        UserFileCleaner // 可选：注销时把磁盘上的上传文件也真删掉
	retention    time.Duration   // 缓冲期（默认 UserPurgeRetention，测试可覆盖）
	retentionSet bool
}

// UserFileCleaner 清理某用户上传的物理文件（由 service.FileService 实现）。
type UserFileCleaner interface {
	RemoveFilesForOwner(ctx context.Context, ownerID string) (int, error)
}

// UserServiceOption 可选依赖/参数：避免为一个可选能力改动全部装配点与测试。
type UserServiceOption func(*UserService)

// WithUserFileCleaner 注入物理文件清理器；未注入时只清台账行、不动磁盘。
func WithUserFileCleaner(c UserFileCleaner) UserServiceOption {
	return func(s *UserService) { s.files = c }
}

// WithPurgeRetention 覆盖缓冲期（测试用；0 表示"立即可清除"）。
func WithPurgeRetention(d time.Duration) UserServiceOption {
	return func(s *UserService) {
		s.retention = d
		s.retentionSet = true
	}
}

func NewUserService(users repository.UserRepository, opts ...UserServiceOption) *UserService {
	s := &UserService{users: users}
	for _, o := range opts {
		o(s)
	}
	return s
}

// purgeRetention 生效的缓冲期。
func (s *UserService) purgeRetention() time.Duration {
	if s.retentionSet {
		return s.retention
	}
	return UserPurgeRetention
}

var (
	// ErrAdminOnly 仅平台管理员可处置账号。
	ErrAdminOnly = errors.New("仅平台管理员可管理用户账号")
	// ErrSuperAdminProtected 内置超管账号不可删除。
	ErrSuperAdminProtected = errors.New("内置超级管理员账号不可删除")
	// ErrUserNotFound 账号不存在（含已注销：注销对普通读取不可见）。
	ErrUserNotFound = errors.New("用户不存在")
)

var (
	// ErrOldPasswordWrong 原密码不正确。
	ErrOldPasswordWrong = errors.New("原密码不正确")
	// ErrWeakPassword 新密码不符合强度要求。
	ErrWeakPassword = errors.New("新密码至少 8 位，且不能与原密码相同")
	// ErrPasswordNotSet 账号从未设置过密码（只能用微信/验证码登录），需管理员重置后再改。
	ErrPasswordNotSet = errors.New("该账号未设置密码，请先由平台管理员重置")
)

const (
	userPasswordMinLen = 8  // 下限：管理员账号不能用弱口令
	userPasswordMaxLen = 72 // bcrypt 只处理前 72 字节，超长直接拒绝而不是静默截断
)

// validateNewPassword 新密码校验：长度 8–72、不与原密码相同、不接受全同字符这类弱口令。
func validateNewPassword(oldPassword, newPassword string) error {
	if len(newPassword) < userPasswordMinLen || len(newPassword) > userPasswordMaxLen {
		return ErrWeakPassword
	}
	if newPassword == oldPassword {
		return ErrWeakPassword
	}
	same := true
	for i := 1; i < len(newPassword); i++ {
		if newPassword[i] != newPassword[0] {
			same = false
			break
		}
	}
	if same {
		return ErrWeakPassword
	}
	return nil
}

// ChangePassword 本人修改密码：校验旧密码 → 写新哈希。
// 仓储在同一事务里自增 token_version 并撤销全部刷新令牌，因此改完必须重新登录。
func (s *UserService) ChangePassword(ctx context.Context, a domain.Actor, oldPassword, newPassword string) error {
	if a.ID == "" {
		return fmt.Errorf("%w: 空 ID", ErrUserNotFound)
	}
	u, err := s.users.FindByID(ctx, a.ID)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrUserNotFound, a.ID)
	}
	if u.PasswordHash == "" {
		return ErrPasswordNotSet
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(oldPassword)) != nil {
		return ErrOldPasswordWrong
	}
	if err := validateNewPassword(oldPassword, newPassword); err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	if err := s.users.UpdatePassword(ctx, a.ID, string(hash)); err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	slog.Info("password changed", "user", a.ID)
	return nil
}

// ResetPassword 平台管理员给其他账号重置密码（无需旧密码；用户随后可用新密码登录）。
func (s *UserService) ResetPassword(ctx context.Context, a domain.Actor, targetID, newPassword string) error {
	if a.Role != domain.RolePlatformAdmin {
		return ErrAdminOnly
	}
	if targetID == "" {
		return fmt.Errorf("%w: 空 ID", ErrUserNotFound)
	}
	if err := validateNewPassword("", newPassword); err != nil {
		return err
	}
	u, err := s.users.FindByID(ctx, targetID)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrUserNotFound, targetID)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	if err := s.users.UpdatePassword(ctx, targetID, string(hash)); err != nil {
		return fmt.Errorf("reset password: %w", err)
	}
	// 审计线索：谁重置了谁、此前有没有密码（不记录任何密码内容）
	slog.Info("password reset by admin", "target", targetID, "admin", a.ID, "had_password", u.PasswordHash != "")
	return nil
}

// UserDeleteResult 删除结果：后台据此如实提示（含缓冲期到期时间与内容处置结果）。
type UserDeleteResult struct {
	ID           string                      `json:"id"`
	Mode         string                      `json:"mode"`          // deleted（唯一动作）
	Status       string                      `json:"status"`        // 删除后的 users.status = deleted
	PurgeAfter   string                      `json:"purge_after"`   // 缓冲期结束时间（到期自动物理清除账号行）
	FilesRemoved int                         `json:"files_removed"` // 磁盘上删除的上传文件数
	KeptContents bool                        `json:"kept_contents"` // 交易/合规记录是否保留（始终保留）
	Cleanup      repository.ContentCleanupReport `json:"cleanup"`  // 内容处置结果（下架/擦除/删除行数）
	CleanupError string                      `json:"cleanup_error,omitempty"` // 清理失败时的原因（账号已失效，内容由缓冲期任务兜底重试）
	Note         string                      `json:"note"`
}

// DeleteUser 删除账号（唯一动作，不可恢复）：
//
//   - 立即失效：status=deleted + token_version 自增 + 回收 refresh_tokens / user_roles；
//   - 从平台消失：FindByID/All 都带 deleted_at IS NULL 过滤，注销后账号不可登录、不进选择器；
//   - 缓冲期：账号行保留 UserPurgeRetention（7 天），到期由 PurgeExpired 后台任务清除；
//   - 内容不删：该用户发布的需求/动态/证书等一律保留（无外键，删账号不影响它们）。
func (s *UserService) DeleteUser(ctx context.Context, a domain.Actor, id string) (UserDeleteResult, error) {
	if a.Role != domain.RolePlatformAdmin {
		return UserDeleteResult{}, ErrAdminOnly
	}
	if id == "" {
		return UserDeleteResult{}, fmt.Errorf("%w: 空 ID", ErrUserNotFound)
	}
	if id == "admin" {
		return UserDeleteResult{}, ErrSuperAdminProtected
	}
	// 不做 FindByID 预检：它带 deleted_at IS NULL 过滤，会把"重复删除"误判成其他错误；
	// 是否存在交给仓储的 RowsAffected 判定（不存在 → repository.ErrUserNotFound）。
	if err := s.users.SoftDelete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return UserDeleteResult{}, fmt.Errorf("%w: %s", ErrUserNotFound, id)
		}
		return UserDeleteResult{}, fmt.Errorf("删除用户 %s: %w", id, err)
	}
	// 账号先失效（硬要求），再处置内容。
	// 顺序要紧：先按台账删磁盘上的物理文件（台账行随后被内容处置删掉，先删行就找不到 key 了），
	// 再执行计划（下架在架内容 + 擦个人信息 + 删纯个人行）。任一步失败都不阻断账号失效，
	// 由缓冲期任务在清除账号行前兜底重试。
	filesRemoved := 0
	if s.files != nil {
		n, ferr := s.files.RemoveFilesForOwner(ctx, id)
		filesRemoved = n
		if ferr != nil {
			slog.Warn("user delete: remove upload files failed", "user", id, "removed", n, "error", ferr)
		}
	}
	clean, cerr := s.users.CleanupUserContent(ctx, id, UserContentCleanupPlan())
	cleanupErr := ""
	if cerr != nil {
		cleanupErr = cerr.Error()
		slog.Warn("user delete: content cleanup failed", "user", id, "error", cerr)
	}
	purgeAfter := time.Now().Add(UserPurgeRetention)
	days := int(UserPurgeRetention.Hours() / 24)
	return UserDeleteResult{
		ID: id, Mode: "deleted", Status: domain.UserDeleted,
		PurgeAfter: purgeAfter.Format(time.RFC3339), KeptContents: true,
		FilesRemoved: filesRemoved,
		Cleanup:      clean, CleanupError: cleanupErr,
		Note: fmt.Sprintf("账号已删除：立即无法登录（旧令牌失效、角色回收），账号行保留 %d 天缓冲期（%s 自动清除，期间不可恢复）。内容处置：下架 %d 条、擦除个人信息 %d 条、删除个人信息 %d 条、删除上传文件 %d 个；工单/合同/资金等履约记录保留",
			days, purgeAfter.Format("2006-01-02"), clean.TakenDown, clean.Wiped, clean.Dropped, filesRemoved),
	}, nil
}

// PurgeExpired 清除超过缓冲期的已注销账号（后台任务调用，返回清除条数）。
// 被工单等外键引用的账号由仓储跳过，下次扫描再试。
func (s *UserService) PurgeExpired(ctx context.Context) (int, error) {
	ids, err := s.users.ListDeletedBefore(ctx, time.Now().Add(-s.purgeRetention()))
	if err != nil {
		return 0, fmt.Errorf("list expired accounts: %w", err)
	}
	plan := UserContentCleanupPlan()
	purged := 0
	for _, id := range ids {
		// 兜底 1：磁盘文件（注销时失败/或账号是本功能上线前删的）
		if s.files != nil {
			if n, ferr := s.files.RemoveFilesForOwner(ctx, id); ferr != nil {
				slog.Warn("user purge: remove upload files failed", "user", id, "removed", n, "error", ferr)
			}
		}
		// 兜底 2：内容处置（幂等）
		if _, err := s.users.CleanupUserContent(ctx, id, plan); err != nil {
			slog.Warn("user purge: content cleanup failed", "user", id, "error", err)
			continue
		}
		if err := s.users.Delete(ctx, id); err != nil {
			if errors.Is(err, repository.ErrUserInUse) {
				continue // 有工单引用：保留账号行，下次扫描再试
			}
			slog.Warn("user purge: delete account failed", "user", id, "error", err)
			continue
		}
		purged++
	}
	return purged, nil
}
