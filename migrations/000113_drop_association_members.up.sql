-- 删除 association_members（协会 8 级角色下线）。
--
-- 背景：这张表与 AssociationRole 的 8 个角色（会长/副会长/秘书长/部门负责人/
-- 普通会员/副会长单位/合作院校/访客）源自一份 .doc 需求，但从未真正启用：
--   · 生产 0 行数据；
--   · 管理端「会员管理」页只是 用户/企业/专家 三个 tab 的容器，前端调
--     /api/v1/association-members 的代码是 0 处；
--   · 唯一读它的判定是 visitorResourceLevel，而 8 个角色里只有 partner
--     （副会长单位）产生过区别——会长和访客在权限上完全等价；
--   · 单条新增接口不校验角色合法性，表列是裸 TEXT（无 CHECK）。
-- 故整块下线，代码侧的实体/仓储/服务/handler 已同步删除。
--
-- ⚠ 与 000111 同款：这是**不可逆**的一步，所以先断言再删。
--   如果表里居然有数据（说明有别处仍在写、或有人手工导入过），
--   直接 RAISE EXCEPTION 让整个迁移失败回滚，而不是把数据删掉。
--   真要在有数据的环境执行，先导出备份，再手工确认后重跑。
DO $$
DECLARE
    remaining INT;
BEGIN
    IF to_regclass('public.association_members') IS NULL THEN
        RETURN;  -- 表不存在（新库或已删过），直接跳过
    END IF;

    EXECUTE 'SELECT count(*) FROM association_members' INTO remaining;

    IF remaining > 0 THEN
        RAISE EXCEPTION 'association_members 还有 % 行数据，拒绝直接删除——请先确认并备份', remaining;
    END IF;
END $$;

DROP TABLE IF EXISTS association_members;
