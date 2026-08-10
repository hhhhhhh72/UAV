-- 回滚：将扩展字段恢复为空值（基础列保留）
UPDATE colleges SET
    city = '', tags = '[]', short_name = '', level_tags = '',
    specialties = '[]', major_count = 0, partner_count = 0, teacher_count = 0, student_count = 0,
    graduate_rate = '', partners = '[]', cover = '', photos = '[]',
    phone = '', website = '', intro = '', majors_detail = '[]'
WHERE id IN ('college-1','college-2','college-3','college-4','college-5','college-6','college-7','college-8');
