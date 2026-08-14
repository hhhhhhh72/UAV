// ============================================================
// 研发难题 · 演示数据（mock）
// ------------------------------------------------------------
// 【接口替换点】后端就绪后：
//   列表：GET /api/v1/rd-challenges  (params: page / page_size)
// 页面已在 list.vue 中处理：接口失败时自动回退到本演示数据。
// 字段遵循后端 snake_case（rd_challenges：title / field /
// description / budget_fen / deadline / status / created_at）。
// ============================================================

export const MOCK_CHALLENGES = [
  {
    id: 'demo-feikong',
    title: '高原环境无人机飞控抗风扰算法研究',
    field: '飞控系统',
    description: '面向海拔 4000m 高原复杂气象环境，研究无人机飞控系统的抗风扰鲁棒控制算法，要求强阵风条件下保持姿态稳定与轨迹精度。',
    budget_fen: 50000000, // ¥50万
    deadline: '2026-08-22T18:00:00+08:00',
    status: 'open',
    created_at: '2026-08-08T10:00:00+08:00',
  },
  {
    id: 'demo-dianchi',
    title: '长航时无人机固态电池能量密度攻关',
    field: '动力电池',
    description: '目标能量密度 ≥ 450Wh/kg，循环寿命 ≥ 800 次，适配长航时工业级无人机平台。',
    budget_fen: 80000000, // ¥80万
    deadline: '2026-08-13T18:00:00+08:00',
    status: 'open',
    created_at: '2026-08-06T10:00:00+08:00',
  },
  {
    id: 'demo-ai',
    title: '低空目标实时检测与避障算法优化',
    field: 'AI算法',
    description: '基于机载算力实现 30ms 内多目标检测与路径重规划，提升低空复杂环境下的自主避障能力。',
    budget_fen: 30000000, // ¥30万
    deadline: '2026-08-30T18:00:00+08:00',
    status: 'open',
    created_at: '2026-08-02T10:00:00+08:00',
  },
  {
    id: 'demo-tongxin',
    title: '城市低空 5G 专网抗干扰链路方案',
    field: '通信链路',
    description: '解决城市多径干扰下的图传 / 数传稳定性问题，保障低空作业链路可靠。',
    budget_fen: 25000000, // ¥25万
    deadline: '2026-09-04T18:00:00+08:00',
    status: 'open',
    created_at: '2026-07-28T10:00:00+08:00',
  },
  {
    id: 'demo-cailiao',
    title: '轻量化碳纤维机身结构工艺优化',
    field: '新型材料',
    description: '整机减重 ≥ 15%，成本下降 ≥ 10%，强度性能不降低，探索低成本量产工艺。',
    budget_fen: 0, // 面议
    deadline: '2026-09-09T18:00:00+08:00',
    status: 'open',
    created_at: '2026-07-25T10:00:00+08:00',
  },
  {
    id: 'demo-zaizai',
    title: '双光云台高精度目标锁定载荷研制',
    field: '载荷设备',
    description: '可见光 + 红外双光联动，锁定精度 ≤ 0.1°，支持低空巡检 / 应急场景。',
    budget_fen: 40000000, // ¥40万
    deadline: '2026-08-15T18:00:00+08:00',
    status: 'open',
    created_at: '2026-07-22T10:00:00+08:00',
  },
]
