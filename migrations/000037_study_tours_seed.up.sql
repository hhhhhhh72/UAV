-- 研学研究：补齐模型字段对应列 + 初始数据（重庆本地低空/无人机研学主题）
ALTER TABLE study_tours ADD COLUMN IF NOT EXISTS description TEXT DEFAULT '';
ALTER TABLE study_tours ADD COLUMN IF NOT EXISTS location TEXT DEFAULT '';
ALTER TABLE study_tours ADD COLUMN IF NOT EXISTS organizer_id TEXT DEFAULT '';
ALTER TABLE study_tours ADD COLUMN IF NOT EXISTS start_date TIMESTAMPTZ;
ALTER TABLE study_tours ADD COLUMN IF NOT EXISTS end_date TIMESTAMPTZ;

INSERT INTO study_tours (id, title, destination, duration, capacity, status, description, location, organizer_id, start_date, end_date, created_at, updated_at) VALUES
('stour-1', '无人机科技科普研学营', '重庆·两江新区', '3天2晚', 30, 'active',
 '走进无人机产业园，了解低空经济发展，动手组装多旋翼无人机并完成试飞，培养青少年科学兴趣。',
 '重庆市两江新区龙兴工业园', 'org-assoc', '2026-08-15T09:00:00+08:00', '2026-08-17T17:00:00+08:00', NOW(), NOW()),
('stour-2', '低空经济产业研学之旅', '重庆·渝北区', '2天1晚', 25, 'active',
 '走进低空经济龙头企业与通航机场，观摩工业级无人机作业演示，对话行业专家，了解产业全链条。',
 '重庆市渝北区空港新城', 'org-assoc', '2026-08-22T09:00:00+08:00', '2026-08-23T17:00:00+08:00', NOW(), NOW()),
('stour-3', '航模制作与飞行体验营', '重庆·巴南区', '1天', 40, 'draft',
 '航模基础知识教学 + 橡皮筋动力飞机/固定翼航模制作 + 户外飞行体验，适合中小学团体。',
 '重庆市巴南区鱼洞体育场', 'org-assoc', '2026-09-05T09:00:00+08:00', '2026-09-05T17:00:00+08:00', NOW(), NOW()),
('stour-4', '航空职业院校开放日研学', '重庆·渝北区', '1天', 50, 'draft',
 '走进无人机应用技术专业院校，体验模拟飞行训练，参观实训基地，了解无人机职业发展方向。',
 '重庆海联职业技术学院', 'org-college', '2026-09-12T09:00:00+08:00', '2026-09-12T16:00:00+08:00', NOW(), NOW()),
('stour-5', '应急救援无人机应用研学', '重庆·北碚区', '2天1晚', 20, 'closed',
 '结合山地救援场景，学习应急无人机侦察/喊话/物资投送应用，参观应急救援演练，感受低空应急协同体系。',
 '重庆市北碚区缙云山', 'org-assoc', '2026-07-18T09:00:00+08:00', '2026-07-19T17:00:00+08:00', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;
