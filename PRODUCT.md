# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

> 微信小程序（uni-app + Vue3，WebView 渲染）—— 按 impeccable 归类为 web；下文原生约束（导航栏/scroll-view/胶囊）为产品约束而非平台分类。

## Stack

现有代码（uni-app + Vue3 `<script setup>`，WXSS rpx 单位）。非新建项目，无需选型。

## Users

- 主用户：成人学员（25-45 岁），为自己考证/职业转型/技能提升报名无人机课程；手机端浏览，常同时对比多家机构后现场决策；关注资质、价格档、能不能报、怎么联系。
- 次要用户：机构/平台管理员（管理侧编辑课程、上传证书图等，学生端不可见管理动作）。

## Product Purpose

让学员在"培训认证 → 课程报名页"完成一次可信的决策：快速看懂这是哪门课、什么价、报什么档位、机构是否被平台认可，然后直接报名或咨询。报名动作跳转 register 表单页；咨询为拨号或意向记录；已满/即将开课明确置灰。

## Positioning

平台背书为主：所有机构课程都经过协会/平台统一审核（资格证、资质、营业信息），页面天然享受"平台认证"的信任——这是机构官网或本地生活平台无法直接复制的机制（协会背景 + 审核资质 + 7 大业务生态）。机构自身实力（评分/通过率/环境）作为第二层证据。

## Operating Context

- 微信小程序主流程（App.vue 微信静默登录）；培训认证 Tab → 课程列表（courses.vue）→ 报名页（enroll.vue）→ 报名表单（register.vue）。
- 页面数据：列表页经 `training_course_detail` storage 传入（无则接口 `/api/v1/training-courses/{id}`）。
- 交互：拨号（makePhoneCall）/ 原生地图（openLocation，无坐标复制地址）/ 收藏（真实接口）/ 详情页价格档经 `training_course_price` 传给 register 回填。
- 状态：recruiting(招生中)/full(已满)/urgent(名额紧张)/upcoming(即将开课)；full/upcoming 报名置灰。

## Capabilities and Constraints

- 已有信息块（必须保留功能）：Hero（图/课程名/机构名/周期/状态徽章）、评分卡、证书类型+机构服务标签、培训参考价（多档选中）、机构简介、联系信息（地址/电话/营业时间）、培训资格证（已认证/待上传，上传仅机构可见）、培训环境、底部栏（收藏/最低价/咨询/报名）。
- 技术约束：uni-app 微信小程序，rpx 单位，禁 emoji 图标（CSS 绘制），品牌色 #0A66C2 / #1DD4A8，`prefers-reduced-motion` 已支持；自定义导航（毛玻璃 hero-nav + 微信胶囊避让）。
- 诚实约束：字段缺失显示"—"/"面议"，不编造默认值（评分/通过率/年限等）。

## Brand Commitments

- 品牌：无人机产业综合服务平台（低空综合服务平台），协会/官方背书属性。
- 色彩承诺：深空蓝 #0A66C2 主色、青绿 #1DD4A8 辅色为全站体系；报名页延续该体系。
- 图标规范：不用 emoji / Unicode 字符图标，一律 CSS 绘制（已用无人机 SVG、日历、定位、手机、时钟等）。

## Evidence on Hand

- 真实数据字段：`courses.vue`/`register.vue` 消费的 detail 对象（title/org_name/prices/courses/status/rating/pass_rate/years/course_types/tags/location/phone/business_hours/certificate/environment/banner/…）。
- 无编造声明：页面缺失字段均显示"—"/"面议"，无模拟报价与虚构评价。

## Product Principles

1. 平台背书先于机构自证：资格证/资质/平台审核是第一条证据线。
2. 五秒决策：一屏内看懂课程名、机构、价格档、状态（可报/已满/即将开课）。
3. 每次可点都是真动作：拨号、导航、收藏、报名、咨询——不允许假成功（toast 冒充提交）。
4. 诚实缺失：没有的数据显示"—"/"面议"，不编造。
5. 成人学员的语气：专业、克制、可信；避免促销泛光与廉价情绪化元素。

## Accessibility & Inclusion

已实现 `prefers-reduced-motion` 全站关闭动画（保留）；小字号（20rpx）与触控目标（76rpx≈38px）需在新设计中提升到 44px（触控下限）并保持正文对比度。无强制 WCAG 标准要求。
