// 课程剩余名额的**单一口径**。
//
// 为什么要有这个文件：课程列表（courses.vue）与课程详情（enroll.vue）此前各写各的 ——
// 列表优先读后端 remain、详情优先读 max_students - enrolled_count。而 remain 曾是建课时
// 由人手填的独立字段（与 max_students 无约束），于是同一门课两个页面给出两个答案：
// 列表「仅剩 10 个」、详情「已报 0 / 11」。
//
// 现在两端都收敛到这里：
//   - 写入口（建课/改课/报名）由后端统一按公式维护 remain；
//   - 读出口统一走本文件，以 max_students - enrolled_count 为准，remain 只作老数据兜底。

// 「名额紧张」阈值：只剩 ≤ 3 个才算。
// 此前的判据是 remain > 0 —— "只要还有名额"就标成紧张（橙色 + 「仅剩 N 个」），
// 一门 30 个名额、0 人报名的课也会显示「仅剩 30 个」，那是制造假稀缺。
export const URGENT_SEATS = 3

// 剩余名额。max_students 缺省或为 0 表示不限额 → 返回 0（调用方按"正常可报名"处理）。
export function remainSeats(item) {
  if (!item) return 0
  const maxStudents = Number(item.max_students)
  if (maxStudents > 0) {
    const enrolled = Number(item.enrolled_count)
    const left = maxStudents - (enrolled > 0 ? enrolled : 0)
    return left > 0 ? left : 0
  }
  const remain = Number(item.remain)
  return remain > 0 ? remain : 0
}

// 是否「名额紧张」：只剩 ≤ URGENT_SEATS 个才算紧张。
export function isSeatsUrgent(item) {
  const left = remainSeats(item)
  return left > 0 && left <= URGENT_SEATS
}
