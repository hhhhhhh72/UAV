ALTER TABLE training_enrollments ALTER COLUMN birthday TYPE TEXT USING COALESCE(birthday::text, '');
ALTER TABLE training_enrollments ALTER COLUMN birthday SET DEFAULT '';
