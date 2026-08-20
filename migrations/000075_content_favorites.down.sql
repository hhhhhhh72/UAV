-- 内容收藏回退：删除三类收藏表（数据不可恢复，仅开发环境回滚用）
DROP TABLE IF EXISTS training_course_favorites;
DROP TABLE IF EXISTS service_listing_favorites;
DROP TABLE IF EXISTS product_favorites;
