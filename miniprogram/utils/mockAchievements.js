// ============================================================
// 科技成果库 / 成果转化 · 演示数据（mock）
// ------------------------------------------------------------
// 【接口替换点】后端就绪后：
//   列表：GET /api/v1/achievements?field=&page=&page_size=
//   详情：GET /api/v1/achievements/{id}
//   转化：GET /api/v1/transformations?achievement_id=
// 页面已在 list.vue / detail.vue / transformations/track.vue 中标注替换位置（搜索“接口替换点”）。
// 字段约定遵循后端 snake_case，前端做优雅降级映射。
// ⚠️ 枚举必须与后端一致：achieve_type / field / stage（lab/pilot/industrialized/listed）。
// ============================================================

// 成果类型（后端 Achievement.achieve_type）
export const ACH_TYPE_LABEL = {
  patent: '发明专利',
  utility: '实用新型',
  copyright: '软件著作',
  paper: '论文成果',
  standard: '技术标准',
  design: '外观设计',
}

// 成果状态（后端 Achievement.status；用于热门/新成果/已转化等运营标签）
export const ACH_STATUS_LABEL = { hot: '热门', transformed: '已转化', new: '新成果' }

// 转化阶段（后端 TransformationStage）：STAGE_SHORT 为展示唯一字面（含 legacy 键别名 laboratory/industrialization）
// STAGE_LABEL 曾带「阶段」后缀与「已产业化」，与展示不一致——现为 STAGE_SHORT 的兼容别名，展示一律以 STAGE_SHORT 为准
export const STAGE_SHORT = { lab: '实验室', laboratory: '实验室', pilot: '中试', industrialization: '产业化', industrialized: '产业化', listed: '已上市' }
export const STAGE_LABEL = STAGE_SHORT
export const STAGE_RANK = { lab: 1, pilot: 2, industrialization: 3, industrialized: 3, listed: 4, launched: 4 }

// 领域 → 图位色（成果库领域映射表：第二调色板成对值，DESIGN.md 2026-08-24 文档轮授权；list/detail 共用同源）
export const FIELD_TONE = {
  '飞控系统': { bg: '#E3EDF9', fg: '#0d47a1' },
  '动力系统': { bg: '#FDEEE4', fg: '#B54708' },
  '载荷设备': { bg: '#FBE9E9', fg: '#b71c1c' },
  '通信链路': { bg: '#E7E9F4', fg: '#1a237e' },
  '集群协同': { bg: '#E5F3F8', fg: '#0e7490' }, // 表格原文
  '遥感测绘': { bg: '#E4F2EF', fg: '#004d40' }, // 借用新型材料青绿（测绘语义）
  'AI算法': { bg: '#F0E9F7', fg: '#4a148c' },
  '地面站': { bg: '#EEF1F4', fg: '#5D6B82' }, // 灰系：站控设备
  '标准规范': { bg: '#EEF1F4', fg: '#344054' }, // 灰系
  // 常见领域补充（线上成果高频字段，缺省曾落到全灰 TONE_DEFAULT）
  '无人机平台': { bg: '#E8F2FC', fg: '#0A66C2' },
  '载荷与传感器': { bg: '#FDF1E7', fg: '#b45309' },
  '导航与定位': { bg: '#E5F0F9', fg: '#1e5eff' },
  '新材料': { bg: '#E4F2EF', fg: '#004d40' },
  '能源动力': { bg: '#FDEEE4', fg: '#B54708' },
}
export const TONE_DEFAULT = { bg: '#EEF1F4', fg: '#344054' }
export const FIELD_ICON = { '飞控系统': '飞', '遥感测绘': '遥', '动力系统': '动', 'AI算法': '算', '载荷设备': '载', '集群协同': '群', '通信链路': '通', '标准规范': '标', '地面站': '地', '无人机平台': '机', '载荷与传感器': '载', '导航与定位': '航', '新材料': '材', '能源动力': '动' }

// 领域 → 封面底色（与现有 FIELD_BG 一致）
export const FIELD_BG = {
  '飞控系统': '#0d47a1', '遥感测绘': '#1b5e20', '动力系统': '#e65100', 'AI算法': '#4a148c',
  '载荷设备': '#b71c1c', '集群协同': '#004d40', '通信链路': '#1a237e', '标准规范': '#37474f', '地面站': '#bf360c',
}

export const MOCK_ACHIEVEMENTS = [
  { id: 'demo-ach-1', owner_id: 'u-1', poster_name: '重庆大学', title: '多旋翼集群协同飞控系统', achieve_type: 'patent', field: '飞控系统', description: '面向多旋翼集群作业场景的协同飞控系统，支持编队组网、避障协同与任务分配，已完成 30kW 级动力平台适配与量产验证，可服务于巡检、植保、应急等集群任务。', stage: 'industrialized', status: 'hot', contact_info: 'tech@cq-uav.org · 023-6789 0000', created_at: '2026-08-01T10:00:00+08:00', updated_at: '2026-08-05T10:00:00+08:00', images: [], attachments: [{ name: '多旋翼集群协同飞控系统技术说明.pdf', size: '2.4 MB', url: '/files/fc.pdf' }, { name: '集群编队演示视频.mp4', size: '45.6 MB', url: '/files/fc.mp4' }] },
  { id: 'demo-ach-2', owner_id: 'u-2', poster_name: '中科院重庆绿色智能研究院', title: '山地遥感测绘载荷平台', achieve_type: 'utility', field: '遥感测绘', description: '适配山地地形的高精度遥感测绘载荷，集成多光谱相机与激光测距，支持仿地飞行与实时建图。', stage: 'pilot', status: 'new', contact_info: '', created_at: '2026-07-18T10:00:00+08:00', updated_at: '2026-07-18T10:00:00+08:00', images: [], attachments: [] },
  { id: 'demo-ach-3', owner_id: 'u-3', poster_name: '宗申航空动力', title: '30kW 航空油电混动动力单元', achieve_type: 'patent', field: '动力系统', description: '面向长航时无人机的油电混动动力单元，兼顾续航与可靠性，支持高原工况。', stage: 'lab', status: '', contact_info: '', created_at: '2026-07-02T10:00:00+08:00', updated_at: '2026-07-02T10:00:00+08:00', images: [], attachments: [] },
  { id: 'demo-ach-4', owner_id: 'u-4', poster_name: '重庆邮电大学', title: '低空航线安全 AI 识别算法', achieve_type: 'copyright', field: 'AI算法', description: '面向低空航线的目标识别与安全预警算法，支持多源感知融合与实时决策。', stage: 'pilot', status: 'hot', contact_info: '', created_at: '2026-06-20T10:00:00+08:00', updated_at: '2026-06-20T10:00:00+08:00', images: [], attachments: [] },
  { id: 'demo-ach-5', owner_id: 'u-5', poster_name: '重庆国飞通用航空', title: '轻量化光电吊舱载荷', achieve_type: 'design', field: '载荷设备', description: '轻量化光电吊舱，适用于侦察、巡检等任务，支持多种载荷快速换装。', stage: 'industrialized', status: '', contact_info: '', created_at: '2026-06-05T10:00:00+08:00', updated_at: '2026-06-05T10:00:00+08:00', images: [], attachments: [] },
  { id: 'demo-ach-6', owner_id: 'u-6', poster_name: '重庆大学', title: '多机编队集群协同算法', achieve_type: 'copyright', field: '集群协同', description: '面向大规模编队的集群协同算法，支持自主避碰与任务分配。', stage: 'lab', status: '', contact_info: '', created_at: '2026-05-22T10:00:00+08:00', updated_at: '2026-05-22T10:00:00+08:00', images: [], attachments: [] },
  { id: 'demo-ach-7', owner_id: 'u-7', poster_name: '重庆电信研究院', title: '5G 低空通信链路方案', achieve_type: 'paper', field: '通信链路', description: '面向低空场景的 5G 通信链路方案，支撑超视距飞行数据回传。', stage: 'lab', status: '', contact_info: '', created_at: '2026-05-08T10:00:00+08:00', updated_at: '2026-05-08T10:00:00+08:00', images: [], attachments: [] },
  { id: 'demo-ach-8', owner_id: 'u-8', poster_name: '重庆市无人机产业协会', title: '无人机巡检作业团体标准', achieve_type: 'standard', field: '标准规范', description: '规范无人机巡检作业流程、数据要求与安全标准，已发布实施。', stage: 'listed', status: 'transformed', contact_info: '', created_at: '2026-04-15T10:00:00+08:00', updated_at: '2026-04-15T10:00:00+08:00', images: [], attachments: [] },
  { id: 'demo-ach-9', owner_id: 'u-9', poster_name: '重庆市勘测院', title: '智能地面站控制系统', achieve_type: 'copyright', field: '地面站', description: '面向多机型统一管控的智能地面站控制系统。', stage: 'industrialized', status: '', contact_info: '', created_at: '2026-03-28T10:00:00+08:00', updated_at: '2026-03-28T10:00:00+08:00', images: [], attachments: [] },
]

// 成果转化演示数据（仅 demo-ach-1）
export const MOCK_TRANSFORMS_BY_ACH = {
  'demo-ach-1': [
    {
      id: 'demo-trans-1',
      achievement_id: 'demo-ach-1',
      owner_id: 'u-1',
      title: '多旋翼集群协同飞控系统产业化',
      stage: 'industrialized',
      progress: '已完成 30kW 动力平台适配与产线验证，正在进行量产爬坡与小批量交付',
      partner_id: '重庆宗申航空发动机',
      status: 'active',
      created_at: '2026-08-01T10:00:00+08:00',
      updated_at: '2026-08-05T10:00:00+08:00',
      milestones: [
        { name: '实验室原型验证通过', completed: true, date: '2025-11-20T00:00:00+08:00', evidence: '' },
        { name: '中试产线建成 · 完成 20 台套', completed: true, date: '2026-03-15T00:00:00+08:00', evidence: '' },
        { name: '30kW 动力平台适配完成', completed: true, date: '2026-06-30T00:00:00+08:00', evidence: '' },
        { name: '量产验证与小批量交付', completed: false, date: '2026-08-05T00:00:00+08:00', evidence: '' },
        { name: '规模上市', completed: false, date: '', evidence: '' },
      ],
    },
  ],
}
