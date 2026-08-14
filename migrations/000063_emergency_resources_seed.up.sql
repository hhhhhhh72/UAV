-- 应急资源台账初始数据（参考资源页设计稿：重庆本地真实场景，不编造伤亡/夸大数据）
-- 注意：contact_info 内手机号为占位演示号码，运营上线前替换为协会真实联系人

INSERT INTO emergency_resources (id, owner_id, name, res_type, specs, quantity, location, contact_info, status, created_at, updated_at) VALUES
('res-drone-1', 'owner-assoc', '应急侦察无人机（M350 RTK）', 'drone', '大疆 M350 RTK ＋ 可见光/热成像双光云台 ＋ RTK 厘米级定位', 6, '重庆市渝中区', '张队长 13800001001', 'available', NOW() - INTERVAL '20 days', NOW() - INTERVAL '20 days'),
('res-drone-2', 'owner-assoc', '双光热成像无人机（M30T）', 'drone', '大疆 M30T ＋ 640×512 热成像 ＋ 激光测距', 4, '重庆市南岸区', '李队 13800001002', 'available', NOW() - INTERVAL '15 days', NOW() - INTERVAL '15 days'),
('res-drone-3', 'owner-assoc', '喊话照明一体无人机', 'drone', '系留式照明 ＋ 100W 喊话器 ＋ 续航 40 分钟', 3, '重庆市北碚区', '王工 13800001003', 'in_use', NOW() - INTERVAL '10 days', NOW() - INTERVAL '2 hours'),
('res-comm-1', 'owner-assoc', '山区图传中继设备', 'comm', '自组网图传中继站 ＋ 覆盖半径 5 公里', 8, '重庆市巴南区', '刘工 13800001004', 'available', NOW() - INTERVAL '12 days', NOW() - INTERVAL '12 days'),
('res-comm-2', 'owner-assoc', '应急对讲指挥终端', 'comm', '数字集群对讲机 ＋ 指挥调度台适配', 20, '重庆市江北区', '陈队 13800001005', 'available', NOW() - INTERVAL '8 days', NOW() - INTERVAL '8 days'),
('res-vehicle-1', 'owner-assoc', '应急指挥车', 'vehicle', '移动指挥方舱 ＋ 图传接收 ＋ 电源保障', 2, '重庆市渝中区', '赵队 13800001006', 'available', NOW() - INTERVAL '30 days', NOW() - INTERVAL '30 days'),
('res-vehicle-2', 'owner-assoc', '无人机运输保障车', 'vehicle', '设备运输 ＋ 充电补给 ＋ 备件存放', 1, '重庆市沙坪坝区', '孙工 13800001007', 'maintenance', NOW() - INTERVAL '25 days', NOW() - INTERVAL '3 days'),
('res-medical-1', 'owner-assoc', '医疗急救包挂载舱', 'medical', '急救物资挂载舱 ＋ 快速投放机构', 10, '重庆市渝北区', '周医生 13800001008', 'available', NOW() - INTERVAL '6 days', NOW() - INTERVAL '6 days'),
('res-medical-2', 'owner-assoc', 'AED 空中投放模块', 'medical', '自动体外除颤器 ＋ 无人机投放挂架', 5, '重庆市九龙坡区', '吴医生 13800001009', 'available', NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days'),
('res-rescue-1', 'owner-assoc', '水域救援抛投装置', 'rescue', '救生圈抛投挂架 ＋ 4 个充气式救生圈', 4, '重庆市北碚区', '郑队 13800001010', 'available', NOW() - INTERVAL '7 days', NOW() - INTERVAL '7 days'),
('res-rescue-2', 'owner-assoc', '山林搜救热成像吊舱', 'rescue', '高灵敏热成像吊舱 ＋ 网格化航线规划', 6, '重庆市南岸区', '冯队 13800001011', 'available', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour')
ON CONFLICT (id) DO NOTHING;
