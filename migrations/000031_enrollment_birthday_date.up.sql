-- 报名生日：TEXT → DATE（纯日期语义，支持日期运算/范围查询）
-- 幂等：仅当列仍为 TEXT 时执行（000030 建列为 TEXT NOT NULL DEFAULT ''，空值兜底 1970-01-01）
-- 注：ALTER TYPE 带 USING 不能用字面 SQL 写在 DO 块（plpgsql 会解析列名），必须 EXECUTE 动态执行
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'training_enrollments' AND column_name = 'birthday' AND data_type = 'text'
  ) THEN
    EXECUTE 'ALTER TABLE training_enrollments ALTER COLUMN birthday DROP DEFAULT';
    EXECUTE 'ALTER TABLE training_enrollments ALTER COLUMN birthday TYPE DATE USING (COALESCE(NULLIF(birthday, ''''), ''1970-01-01'')::date)';
  END IF;
END $$;
