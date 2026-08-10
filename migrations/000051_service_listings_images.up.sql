-- 服务能力封面图（复用小程序主题图，镜像位于后端 uploads/ 目录）
-- sl-solar.jpg 光伏巡检 / sl-lift.jpg 吊运作业 / sl-hero.jpg 飞行巡检
UPDATE service_listings SET image = '/uploads/sl-solar.jpg' WHERE id = 'sl-1';
UPDATE service_listings SET image = '/uploads/sl-hero.jpg' WHERE id = 'sl-2';
UPDATE service_listings SET image = '/uploads/sl-lift.jpg' WHERE id = 'sl-3';
UPDATE service_listings SET image = '/uploads/sl-hero.jpg' WHERE id = 'sl-4';
UPDATE service_listings SET image = '/uploads/sl-solar.jpg' WHERE id = 'sl-5';
UPDATE service_listings SET image = '/uploads/sl-lift.jpg' WHERE id = 'sl-6';
