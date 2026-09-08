-- 报名表补身份证反面影像列（白底证件照 + 身份证正反面 3 张材料闭环）
ALTER TABLE training_enrollments ADD COLUMN id_card_back varchar(500) NOT NULL DEFAULT '';
