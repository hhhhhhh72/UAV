# -*- coding: utf-8 -*-
"""替换 PRD 树形架构图段落为 7 大业务系统结构"""
import sys
sys.stdout.reconfigure(encoding='utf-8')
from docx import Document
from docx.oxml.ns import qn
from docx.oxml import OxmlElement

PATH = r'D:/w-yao/docs/需求文档/V1-PRD-低空经济生态服务平台.docx'
doc = Document(PATH)
body = doc.element.body

lines = [
    '低空经济生态服务平台（7大业务系统 + 通用支撑）',
    '│',
    '├── ① 会员生态资源管控',
    '│   ├── 1. 用户与权限管理（4级RBAC + 协会7级细分）',
    '│   ├── 2. 会员单位管理（企业档案/审核/台账）',
    '│   ├── 3. 专家智库',
    '│   ├── 4. 产业资源台账（场地/试飞/中试）',
    '│   └── 5. 人才资源库',
    '│',
    '├── ② 产业供需智能对接',
    '│   ├── 6. 作业需求大厅（发布/智能匹配/竞标）',
    '│   ├── 7. 供给能力展示（整机/零部件/服务/场地）',
    '│   ├── 8. 接单与派单流程',
    '│   └── 9. 订单管理与评价',
    '│',
    '├── ③ 产学研协同创新',
    '│   ├── 10. 技术成果库',
    '│   ├── 11. 研发难题广场',
    '│   ├── 12. 课题联合攻关',
    '│   ├── 13. 中试基地测试预约',
    '│   └── 14. 成果转化追踪',
    '│',
    '├── ④ 合规政策服务',
    '│   ├── 15. 资讯与政策中心',
    '│   ├── 16. 合规知识库',
    '│   ├── 17. 团体标准管理',
    '│   ├── 18. 项目申报服务',
    '│   └── 19. 企业案例库',
    '│',
    '├── ⑤ 人才教育与产教融合',
    '│   ├── 20. 培训课程管理（课程超市）',
    '│   ├── 21. 培训报名与排班',
    '│   ├── 22. 考证管理（证书/执照）',
    '│   ├── 23. 赛事管理',
    '│   ├── 24. 招聘求职',
    '│   └── 25. 院校展示与校企共建',
    '│',
    '├── ⑥ 活动与品牌服务',
    '│   ├── 26. 活动管理',
    '│   ├── 27. 会员品牌展示',
    '│   ├── 28. 展会排期',
    '│   └── 29. 行业报告发布',
    '│',
    '├── ⑦ 低空应急资源协同',
    '│   ├── 30. 应急资源建档',
    '│   ├── 31. 一键调度',
    '│   ├── 32. 救援案例库',
    '│   ├── 33. 部门对接',
    '│   └── 34. 联合演练',
    '│',
    '└── 【通用支撑】',
    '    ├── 35. 认证授权（微信登录/RBAC/刷新令牌）',
    '    ├── 36. 消息通知（站内消息/已读未读）',
    '    ├── 37. 社区内容（发帖/评论/举报）',
    '    ├── 38. 二手交易（商品发布/收藏）',
    '    ├── 39. 用工派遣（用工订单/报价）',
    '    ├── 40. 合同签约（模板/签章回调/作废）',
    '    ├── 41. 资金托管（充值/冻结/释放/退款）',
    '    ├── 42. 文件上传（文件存储/图片优化）',
    '    └── 43. 数据统计看板（管理后台/CSV导出）',
]

# 定位树形段落
tree_p = None
for child in body.iterchildren():
    if child.tag == qn('w:p'):
        texts = ''.join(n.text or '' for n in child.iter(qn('w:t')))
        if texts.startswith('低空经济生态服务平台 V1'):
            tree_p = child
            break

if tree_p is None:
    print('FAIL: 未找到树形段落')
    sys.exit(1)

# 保留 pPr，删除所有 run，重建多行
for r in list(tree_p.findall(qn('w:r'))):
    tree_p.remove(r)

for i, line in enumerate(lines):
    r = OxmlElement('w:r')
    rPr = OxmlElement('w:rPr')
    rFonts = OxmlElement('w:rFonts')
    rFonts.set(qn('w:eastAsia'), '宋体')
    rPr.append(rFonts)
    r.append(rPr)
    t = OxmlElement('w:t')
    t.set(qn('xml:space'), 'preserve')
    t.text = line
    r.append(t)
    tree_p.append(r)
    if i < len(lines) - 1:
        br = OxmlElement('w:br')
        tree_p.append(br)

doc.save(PATH)
print('OK: 树形段落已替换为 7 大系统架构图（44 行）')
