-- 赛事活动初始数据（小程序赛事列表/详情/报名页展示，字段对齐 000044 扩展列）
INSERT INTO competitions (
    id, title, category, description, location, sponsor, start_date, end_date, deadline,
    max_teams, reg_count, organizer_sub, fee, min_fee, tags, poster, registration_status,
    requirements, events, prizes, status, created_at, updated_at
) VALUES
('comp-1', '2026全国无人机职业技能大赛', '多旋翼',
 '面向全国无人机从业者的职业技能竞赛，涵盖多旋翼巡检、植保作业、航拍测绘等实操赛项，优胜者可获行业认证。',
 '深圳宝安区国际会展中心', '中国航空器拥有者及驾驶员协会',
 '2026-09-15 09:00:00+08', '2026-09-18 18:00:00+08', '2026-09-01 23:59:59+08',
 200, 128, '深圳市低空经济产业协会', 380, 380,
 '["多旋翼","固定翼","国家级"]', '/static/home/hero-inspection.jpg', 'open',
 '[
   {"icon":"✈","name":"多旋翼巡检","desc":"电力线路巡检实操","level":"高级"},
   {"icon":"🌾","name":"植保作业","desc":"农用植保实操","level":"中级"},
   {"icon":"📷","name":"航拍测绘","desc":"正射影像采集与建模","level":"高级"}
 ]',
 '[
   {"name":"预选赛","type":"个人赛","format":"实操","fee":0},
   {"name":"决赛","type":"团体赛","format":"实操+答辩","fee":380}
 ]',
 '[
   {"level":"一等奖","amount":30000,"metal":"gold","medal":"金牌"},
   {"level":"二等奖","amount":15000,"metal":"silver","medal":"银牌"},
   {"level":"三等奖","amount":8000,"metal":"bronze","medal":"铜牌"}
 ]',
 'enrolling', NOW(), NOW()),
('comp-2', '首届西南无人机FPV竞速挑战赛', '竞速FPV',
 '西南地区首个无人机 FPV 竞速赛事，设置多旋翼竞速、障碍穿越等赛项，面向青少年与职业飞手开放。',
 '成都天府新区无人机竞速基地', '四川省航空运动协会',
 '2026-10-01 09:00:00+08', '2026-10-03 18:00:00+08', '2026-09-20 23:59:59+08',
 100, 56, '成都市无人机产业协会', 280, 280,
 '["竞速FPV","多旋翼"]', '/static/home/demand-lift.jpg', 'open',
 '[
   {"icon":"🚁","name":"竞速障碍穿越","desc":"多旋翼竞速穿越障碍","level":"中级"},
   {"icon":"🏁","name":"FPV竞速","desc":"FPV 竞速计时赛","level":"高级"}
 ]',
 '[
   {"name":"小组赛","type":"个人赛","format":"竞速","fee":0},
   {"name":"淘汰赛","type":"个人赛","format":"竞速","fee":280}
 ]',
 '[
   {"level":"冠军","amount":10000,"metal":"gold","medal":"奖杯"},
   {"level":"亚军","amount":5000,"metal":"silver","medal":"银牌"},
   {"level":"季军","amount":3000,"metal":"bronze","medal":"铜牌"}
 ]',
 'enrolling', NOW(), NOW()),
('comp-3', '2026无人机创新应用大赛', '创新应用',
 '面向高校与企业团队的无人机创新应用大赛，聚焦低空经济新场景、新技术，设置方案评审与路演答辩环节。',
 '北京亦庄经济技术开发区', '工信部人才交流中心',
 '2026-08-01 09:00:00+08', '2026-08-15 18:00:00+08', NULL,
 150, 256, '北京市低空经济促进会', 0, 0,
 '["航拍","固定翼","国家级"]', '', 'closed',
 '[
   {"icon":"💡","name":"创新方案","desc":"无人机创新应用方案设计","level":"不限"},
   {"icon":"🎤","name":"路演答辩","desc":"现场路演与专家答辩","level":"不限"}
 ]',
 '[
   {"name":"初赛","type":"团体赛","format":"方案评审","fee":0},
   {"name":"决赛","type":"团体赛","format":"路演答辩","fee":0}
 ]',
 '[
   {"level":"金奖","amount":50000,"metal":"gold","medal":"奖杯"},
   {"level":"银奖","amount":30000,"metal":"silver","medal":"银牌"},
   {"level":"铜奖","amount":10000,"metal":"bronze","medal":"铜牌"}
 ]',
 'ongoing', NOW(), NOW()),
('comp-4', '青少年无人机编程挑战赛', '编程挑战',
 '面向中小学生群体的无人机编程挑战赛，通过图形化编程与无人机实操结合，培养青少年科学素养。',
 '上海市浦东新区青少年活动中心', '上海市教委',
 '2026-11-01 09:00:00+08', '2026-11-02 18:00:00+08', '2026-10-20 23:59:59+08',
 300, 340, '上海市青少年科技教育中心', 120, 120,
 '["多旋翼","航拍","青少年"]', '/static/home/demand-solar.jpg', 'open',
 '[
   {"icon":"🧑‍🔬","name":"编程挑战","desc":"图形化编程任务","level":"初级"},
   {"icon":"✈","name":"飞行实操","desc":"编程飞行任务实操","level":"初级"}
 ]',
 '[
   {"name":"预赛","type":"个人赛","format":"编程任务","fee":0},
   {"name":"决赛","type":"个人赛","format":"编程+实操","fee":120}
 ]',
 '[
   {"level":"一等奖","amount":5000,"metal":"gold","medal":"金牌"},
   {"level":"二等奖","amount":3000,"metal":"silver","medal":"银牌"},
   {"level":"三等奖","amount":1000,"metal":"bronze","medal":"铜牌"}
 ]',
 'enrolling', NOW(), NOW()),
('comp-5', '国际无人机系统博览会竞技赛', '综合竞技',
 '国际无人机系统博览会配套竞技赛事，汇聚国内外顶尖飞手与行业组织，设固定翼、多旋翼、行业应用三大类目。',
 '广州琶洲国际会展中心', '广州市低空经济产业协会',
 '2026-12-05 09:00:00+08', '2026-12-07 18:00:00+08', '2026-11-20 23:59:59+08',
 500, 89, '广东省无人机行业协会', 580, 580,
 '["多旋翼","固定翼","国际赛"]', '/static/home/home-bg.jpg', 'open',
 '[
   {"icon":"🌐","name":"国际组","desc":"国际飞手公开组","level":"高级"},
   {"icon":"🏭","name":"行业应用","desc":"行业应用场景赛","level":"中级"}
 ]',
 '[
   {"name":"公开组","type":"个人赛","format":"综合","fee":580},
   {"name":"行业组","type":"团体赛","format":"任务实操","fee":580}
 ]',
 '[
   {"level":"冠军","amount":80000,"metal":"gold","medal":"奖杯"},
   {"level":"亚军","amount":40000,"metal":"silver","medal":"银牌"},
   {"level":"季军","amount":20000,"metal":"bronze","medal":"铜牌"}
 ]',
 'enrolling', NOW(), NOW()),
('comp-6', '2026贵州无人机应急救援演练赛', '应急演练',
 '联合应急管理部门举办的无人机应急救援演练赛，模拟自然灾害场景下的无人机搜救、物资投送等科目。',
 '贵阳市观山湖区应急指挥中心', '贵州省应急管理厅',
 '2026-06-10 09:00:00+08', '2026-06-12 18:00:00+08', '2026-05-25 23:59:59+08',
 80, 72, '贵州省无人机产业协会', 0, 0,
 '["多旋翼","航拍","应急"]', '', 'closed',
 '[
   {"icon":"🚨","name":"搜救科目","desc":"无人机搜救实操","level":"高级"},
   {"icon":"📦","name":"物资投送","desc":"紧急物资定点投送","level":"高级"}
 ]',
 '[
   {"name":"演练","type":"团体赛","format":"模拟演练","fee":0}
 ]',
 '[
   {"level":"优秀团队","amount":0,"metal":"","medal":"荣誉证书"},
   {"level":"优秀飞手","amount":0,"metal":"","medal":"荣誉证书"}
 ]',
 'closed', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;
