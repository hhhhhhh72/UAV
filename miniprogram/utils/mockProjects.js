// ============================================================
// 课题攻关 · 演示数据（mock）
// ------------------------------------------------------------
// 【接口替换点】后端就绪后：
//   列表：GET /api/v1/research-projects  (params: page / page_size)
// 页面已在 list.vue 中处理：接口失败时自动回退到本演示数据。
// 字段遵循后端 snake_case（research_projects：title / field /
// description / lead_org / members / budget_fen / start_date /
// end_date / milestones / status / created_at）。
// ============================================================

export const MOCK_PROJECTS = [
  {
    id: 'demo-feikong-jg',
    title: '高原环境无人机飞控抗风扰鲁棒控制研究',
    field: '飞控系统',
    description: '面向海拔 4000m 高原复杂气象环境，研究飞控系统抗风扰鲁棒控制算法，强阵风条件下保持姿态稳定与轨迹精度。',
    lead_org: '重庆大学',
    members: ['重庆大学航空航天学院', '重庆华科尔无人机', '重庆两江航空产业投资集团'],
    budget_fen: 80000000, // ¥80万
    start_date: '2026-07-01T00:00:00+08:00',
    end_date: '2026-08-16T18:00:00+08:00',
    milestones: '3 个月出仿真验证报告，6 个月试飞验收',
    status: 'recruiting',
    created_at: '2026-07-28T10:00:00+08:00',
  },
  {
    id: 'demo-dianchi-jg',
    title: '长航时无人机固态电池装机验证',
    field: '动力电池',
    description: '固态电池能量密度 ≥ 450Wh/kg 装机验证，循环寿命 ≥ 800 次，适配长航时工业级平台。',
    lead_org: '重庆理工大学',
    members: ['重庆理工大学材料学院', '重庆蓝海新能源'],
    budget_fen: 120000000, // ¥120万
    start_date: '2026-06-15T00:00:00+08:00',
    end_date: '2026-11-30T18:00:00+08:00',
    milestones: '9 月完成电池包集成，11 月整机验证',
    status: 'progress',
    created_at: '2026-07-20T10:00:00+08:00',
  },
  {
    id: 'demo-ai-jg',
    title: '低空多目标实时检测与机载算力部署',
    field: 'AI算法',
    description: '基于机载边缘算力实现 30ms 内多目标检测与路径重规划，交付可部署模型与推理框架。',
    lead_org: '电子科技大学',
    members: ['电子科技大学自动化学院', '重庆信通研究院'],
    budget_fen: 0, // 面议
    start_date: '2026-07-10T00:00:00+08:00',
    end_date: '2026-09-20T18:00:00+08:00',
    milestones: '8 月完成算法选型，9 月机载部署验证',
    status: 'recruiting',
    created_at: '2026-07-15T10:00:00+08:00',
  },
  {
    id: 'demo-tongxin-jg',
    title: '城市低空 5G 专网抗干扰链路组网方案',
    field: '通信链路',
    description: '解决城市多径干扰下图传/数传稳定性问题，形成低空作业通信链路组网技术方案。',
    lead_org: '重庆邮电大学',
    members: ['重庆邮电大学通信学院', '重庆移动 5G 联合实验室'],
    budget_fen: 60000000, // ¥60万
    start_date: '2026-09-01T00:00:00+08:00',
    end_date: '2026-10-31T18:00:00+08:00',
    milestones: '10 月输出组网方案与测试报告',
    status: 'planning',
    created_at: '2026-08-01T10:00:00+08:00',
  },
  {
    id: 'demo-cailiao-jg',
    title: '轻量化碳纤维机身结构工艺攻关',
    field: '新型材料',
    description: '整机减重 ≥ 15%、成本下降 ≥ 10%，强度不降低，形成低成本量产工艺路线。',
    lead_org: '四川大学',
    members: ['四川大学高分子学院', '重庆宗申航空动力'],
    budget_fen: 200000000, // ¥200万
    start_date: '2026-05-20T00:00:00+08:00',
    end_date: '2026-12-15T18:00:00+08:00',
    milestones: '10 月工艺定型，12 月量产试制',
    status: 'progress',
    created_at: '2026-07-08T10:00:00+08:00',
  },
  {
    id: 'demo-zaizai-jg',
    title: '双光云台高精度目标锁定载荷研制',
    field: '载荷设备',
    description: '可见光 + 红外双光联动，锁定精度 ≤ 0.1°，完成样机定型与小批量试产。',
    lead_org: '中国电科 44 所',
    members: ['中国电科 44 所', '重庆驰宇光电'],
    budget_fen: 50000000, // ¥50万
    start_date: '2026-03-01T00:00:00+08:00',
    end_date: '2026-06-30T18:00:00+08:00',
    milestones: '已完成样机联调与试飞验收',
    status: 'completed',
    created_at: '2026-06-20T10:00:00+08:00',
  },
  {
    id: 'demo-biaozhun-jg',
    title: '低空无人机适航审定团体标准编制',
    field: '技术标准',
    description: '牵头编制低空无人机适航审定团体标准，覆盖运行安全、数据链路与应急回收要求。',
    lead_org: '重庆无人机产业协会',
    members: ['重庆无人机产业协会', '中国民航大学'],
    budget_fen: 30000000, // ¥30万
    start_date: '2026-09-15T00:00:00+08:00',
    end_date: '2027-03-31T18:00:00+08:00',
    milestones: '12 月完成征求意见稿，次年 3 月发布',
    status: 'planning',
    created_at: '2026-08-05T10:00:00+08:00',
  },
  {
    id: 'demo-jiqun-jg',
    title: '集群编队协同作业控制平台',
    field: '集群协同',
    description: '10 架级无人机集群编队、任务分配与协同避障，形成可复用的集群作业控制平台。',
    lead_org: '重庆大学',
    members: ['重庆大学自动化学院', '重庆翼动科技', '重庆两江协同创新区'],
    budget_fen: 150000000, // ¥150万
    start_date: '2026-07-05T00:00:00+08:00',
    end_date: '2026-08-19T18:00:00+08:00',
    milestones: '8 月完成 10 机编队飞行验证',
    status: 'recruiting',
    created_at: '2026-08-08T10:00:00+08:00',
  },
]
