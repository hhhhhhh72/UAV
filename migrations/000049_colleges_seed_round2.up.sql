-- 院校扩展字段补全（对齐 000044 新增列：city/tags/short_name/level_tags/specialties/major_count/partner_count/teacher_count/student_count/graduate_rate/partners/cover/photos/phone/website/intro/majors_detail）
-- 注：majors/facilities 基础列已在 000034 写入，此处只补扩展列；city 与 region 保持一致
UPDATE colleges SET
    city = '北京', tags = '["985","211"]', short_name = '北航', level_tags = '985/211',
    specialties = '["飞行器设计","无人机系统","航空航天工程"]',
    major_count = 6, partner_count = 120, teacher_count = 480, student_count = 32000,
    graduate_rate = '92.5%',
    partners = '[{"icon":"商","name":"大疆创新","type":"共建实验室"},{"icon":"航","name":"航天科工","type":"产学研合作"},{"icon":"翼","name":"中航工业","type":"定向培养"}]',
    cover = '', photos = '[]',
    phone = '010-82317670', website = 'www.buaa.edu.cn',
    intro = '航空科学与工程学院是国内顶尖的航空航天教育基地，在无人机系统设计、飞行控制与导航领域拥有国家重点实验室支撑，与平台共建低空经济人才培养基地。',
    majors_detail = '[
      {"name":"飞行器设计与工程","degree":"本科","duration":"4年","key":true,"flagship":true},
      {"name":"无人驾驶航空器系统工程","degree":"本科","duration":"4年","key":true,"flagship":true},
      {"name":"控制科学与工程","degree":"硕士","duration":"3年","key":false,"flagship":false}
    ]'
WHERE id = 'college-1';

UPDATE colleges SET
    city = '南京', tags = '["985","211"]', short_name = '南航', level_tags = '985/211',
    specialties = '["无人机应用","适航技术","飞行器设计"]',
    major_count = 5, partner_count = 90, teacher_count = 420, student_count = 28000,
    graduate_rate = '93.1%',
    partners = '[{"icon":"江","name":"南京航空航天大学科技园","type":"产学研基地"},{"icon":"翼","name":"亿航智能","type":"共建实验室"}]',
    cover = '', photos = '[]',
    phone = '025-84892527', website = 'www.nuaa.edu.cn',
    intro = '民航学院是首批设立无人机应用技术专业的高校之一，设有民航适航实验室与无人机飞行试验基地，与多家无人机企业共建产学研基地。',
    majors_detail = '[
      {"name":"飞行器适航技术","degree":"本科","duration":"4年","key":true,"flagship":true},
      {"name":"无人机应用技术","degree":"本科","duration":"4年","key":true,"flagship":false},
      {"name":"飞行器设计与工程","degree":"硕士","duration":"3年","key":false,"flagship":false}
    ]'
WHERE id = 'college-2';

UPDATE colleges SET
    city = '西安', tags = '["985","211"]', short_name = '西工大', level_tags = '985/211',
    specialties = '["飞行控制","无人机系统","飞行器动力"]',
    major_count = 7, partner_count = 150, teacher_count = 550, student_count = 36000,
    graduate_rate = '94.0%',
    partners = '[{"icon":"航","name":"中国飞行试验研究院","type":"联合试验"},{"icon":"翼","name":"中航无人机","type":"定向培养"}]',
    cover = '', photos = '[]',
    phone = '029-88493000', website = 'www.nwpu.edu.cn',
    intro = '无人机特种技术重点实验室依托单位，在军用和民用无人机领域均有深厚的研究积累，承担多项国家级无人机关键技术攻关项目。',
    majors_detail = '[
      {"name":"无人系统科学与技术","degree":"本科","duration":"4年","key":true,"flagship":true},
      {"name":"飞行器动力工程","degree":"本科","duration":"4年","key":false,"flagship":false},
      {"name":"控制科学与工程","degree":"硕士","duration":"3年","key":true,"flagship":false}
    ]'
WHERE id = 'college-3';

UPDATE colleges SET
    city = '成都', tags = '["专科","高职"]', short_name = '成都航院', level_tags = '专科/高职',
    specialties = '["无人机装调","航拍测绘","农业植保"]',
    major_count = 4, partner_count = 40, teacher_count = 180, student_count = 14000,
    graduate_rate = '95.8%',
    partners = '[{"icon":"成","name":"成飞集团","type":"校企共建"},{"icon":"发","name":"成都飞机设计研究所","type":"实习基地"}]',
    cover = '', photos = '[]',
    phone = '028-88459322', website = 'www.cap.edu.cn',
    intro = '西南地区最早开设无人机应用技术专业的高职院校，与成飞、成发等企业深度合作，实训条件完善，毕业生供不应求。',
    majors_detail = '[
      {"name":"无人机应用技术","degree":"专科","duration":"3年","key":true,"flagship":true},
      {"name":"无人机装调与维修","degree":"专科","duration":"3年","key":true,"flagship":true},
      {"name":"航拍测绘技术","degree":"专科","duration":"3年","key":false,"flagship":false}
    ]'
WHERE id = 'college-4';

UPDATE colleges SET
    city = '长沙', tags = '["专科","高职"]', short_name = '长沙航院', level_tags = '专科/高职',
    specialties = '["无人机装调","农业植保","电力巡检"]',
    major_count = 3, partner_count = 35, teacher_count = 160, student_count = 12000,
    graduate_rate = '96.2%',
    partners = '[{"icon":"中","name":"中航工业","type":"校企共建"},{"icon":"发","name":"中国航发","type":"实训基地"}]',
    cover = '', photos = '[]',
    phone = '0731-85473700', website = 'www.cavtc.cn',
    intro = '与中航工业、中国航发等企业共建实训基地，注重实操能力培养，无人机装调与植保作业实训场全国领先。',
    majors_detail = '[
      {"name":"无人机应用技术","degree":"专科","duration":"3年","key":true,"flagship":true},
      {"name":"农业植保无人机","degree":"专科","duration":"3年","key":true,"flagship":true}
    ]'
WHERE id = 'college-5';

UPDATE colleges SET
    city = '天津', tags = '["本科","双一流"]', short_name = '中国民航大学', level_tags = '本科/双一流',
    specialties = '["适航管理","无人机运行","飞行技术"]',
    major_count = 5, partner_count = 80, teacher_count = 300, student_count = 25000,
    graduate_rate = '91.8%',
    partners = '[{"icon":"民","name":"民航局适航审定中心","type":"共建实验室"},{"icon":"航","name":"大疆行业应用","type":"产学研合作"}]',
    cover = '', photos = '[]',
    phone = '022-24092626', website = 'www.cauc.edu.cn',
    intro = '民航系统唯一博士学位授予单位，设有无人机适航与运行管理专业方向，为民航局与行业输出适航审定人才。',
    majors_detail = '[
      {"name":"无人机适航与运行管理","degree":"本科","duration":"4年","key":true,"flagship":true},
      {"name":"飞行技术","degree":"本科","duration":"4年","key":false,"flagship":false},
      {"name":"适航技术与管理","degree":"硕士","duration":"3年","key":true,"flagship":false}
    ]'
WHERE id = 'college-6';

UPDATE colleges SET
    city = '重庆', tags = '["专科","高职"]', short_name = '重庆海联职院', level_tags = '专科/高职',
    specialties = '["无人机应用技术","低空运维","航拍测绘"]',
    major_count = 3, partner_count = 25, teacher_count = 120, student_count = 9000,
    graduate_rate = '94.5%',
    partners = '[{"icon":"渝","name":"重庆无人机产业协会","type":"共建实训基地"},{"icon":"翼","name":"重庆翼航科技","type":"定向输送"}]',
    cover = '', photos = '[]',
    phone = '023-67198000', website = 'www.hailian.com',
    intro = '重庆市无人机产业协会合作院校，开设无人机应用技术专业，建有无人机实训基地与飞行训练场，为平台会员企业输送飞手与运维人才。',
    majors_detail = '[
      {"name":"无人机应用技术","degree":"专科","duration":"3年","key":true,"flagship":true},
      {"name":"低空运维管理","degree":"专科","duration":"3年","key":true,"flagship":true}
    ]'
WHERE id = 'college-7';

UPDATE colleges SET
    city = '重庆', tags = '["专科","高职"]', short_name = '重庆交通职院', level_tags = '专科/高职',
    specialties = '["无人机应用","交通运维","智能装备"]',
    major_count = 3, partner_count = 20, teacher_count = 130, student_count = 10000,
    graduate_rate = '95.0%',
    partners = '[{"icon":"渝","name":"重庆无人机产业协会","type":"校企共建订单班"},{"icon":"交","name":"重庆交运集团","type":"定向培养"}]',
    cover = '', photos = '[]',
    phone = '023-47287000', website = 'www.cqjzc.edu.cn',
    intro = '重庆本地高职院校，与协会共建校企合作订单班，聚焦无人机应用与交通运维方向，实训设施完善。',
    majors_detail = '[
      {"name":"无人机应用技术","degree":"专科","duration":"3年","key":true,"flagship":true},
      {"name":"智能装备运维","degree":"专科","duration":"3年","key":false,"flagship":false}
    ]'
WHERE id = 'college-8';
