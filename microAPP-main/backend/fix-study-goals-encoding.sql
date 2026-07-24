-- 修复研学目标中的乱码数据
-- 请在执行前备份数据库

UPDATE service_configs 
SET config = jsonb_set(
  config::jsonb,
  '{9,packages,study-halfday,studyGoals,1,content}',
  '"通过动手组装木质无人机与飞行实践，培养学生的动手能力、逻辑思维、空间判断能力与团队协作意识。"'
)
WHERE config::jsonb->'9'->'packages'->'study-halfday'->'studyGoals'->1->>'content' LIKE '%木质%';

-- 如果有其他课程包也有类似问题，请分别执行
-- study-family 课程包
UPDATE service_configs 
SET config = jsonb_set(
  config::jsonb,
  '{9,packages,study-family,studyGoals,1,content}',
  '"通过动手组装木质无人机与飞行实践，培养学生的动手能力、逻辑思维、空间判断能力与团队协作意识。"'
)
WHERE config::jsonb->'9'->'packages'->'study-family'->'studyGoals'->1->>'content' LIKE '%木质%';
