/**
 * 管理相关路由
 */
const express = require('express');
const bcrypt = require('bcryptjs');
const { randomUUID } = require('crypto');
const XLSX = require('xlsx');
const { config } = require('../config');
const { logger } = require('../logger');
const {
  readUsersDB, writeUsersDB,
  readApplicationsDB, writeApplicationsDB,
  readCasesDB, writeCasesDB,
  readServicesConfig, writeServicesConfig,
  readReviewsDB, writeReviewsDB
} = require('../storage');
const { authRequired, roleRequired } = require('../middleware/auth');
const { sanitizeBody, requireFields, validateQuery } = require('../middleware/validation');
const { asyncHandler } = require('../middleware/error');

const router = express.Router();

// 所有管理路由都需要认证
router.use(authRequired);

/**
 * 获取仪表板统计数据
 */
router.get('/stats', asyncHandler(async (req, res) => {
  const [applications, users, cases] = await Promise.all([
    readApplicationsDB(),
    readUsersDB(),
    readCasesDB()
  ]);

  let filteredApplications = applications;

  // 根据角色过滤数据
  if (req.user.role === 'dsl_admin') {
    filteredApplications = applications.filter(app => app.serviceId === '13');
  } else if (req.user.role === 'study_admin') {
    filteredApplications = applications.filter(app => app.serviceId === '9');
  }

  const stats = {
    totalOrders: filteredApplications.length,
    pendingOrders: filteredApplications.filter(app => app.status === '待处理').length,
    processingOrders: filteredApplications.filter(app => app.status === '处理中').length,
    completedOrders: filteredApplications.filter(app => app.status === '已完成').length,
    totalUsers: users.length,
    totalCases: cases.length,
    recentOrders: filteredApplications
      .sort((a, b) => new Date(b.createTime) - new Date(a.createTime))
      .slice(0, 10)
  };

  res.json({
    success: true,
    stats
  });
}));

/**
 * 获取申请列表
 */
router.get('/applications', validateQuery({
  page: { type: 'number', min: 1 },
  limit: { type: 'number', min: 1, max: 100 },
  status: { type: 'string' },
  serviceId: { type: 'string' }
}), asyncHandler(async (req, res) => {
  const { page = 1, limit = 20, status, serviceId } = req.query;

  let applications = await readApplicationsDB();

  // 根据角色过滤
  if (req.user.role === 'dsl_admin') {
    applications = applications.filter(app => app.serviceId === '13');
  } else if (req.user.role === 'study_admin') {
    applications = applications.filter(app => app.serviceId === '9');
  }

  // 状态筛选
  if (status) {
    applications = applications.filter(app => app.status === status);
  }

  // 服务ID筛选
  if (serviceId) {
    applications = applications.filter(app => app.serviceId === serviceId);
  }

  // 排序
  applications.sort((a, b) => new Date(b.createTime) - new Date(a.createTime));

  // 分页
  const total = applications.length;
  const start = (page - 1) * limit;
  const paginated = applications.slice(start, start + limit);

  res.json({
    success: true,
    data: paginated,
    pagination: {
      page: parseInt(page),
      limit: parseInt(limit),
      total,
      totalPages: Math.ceil(total / limit)
    }
  });
}));

/**
 * 更新申请状态
 */
router.post('/applications/:id', sanitizeBody, asyncHandler(async (req, res) => {
  const { id } = req.params;
  const { status, remark } = req.body;

  const applications = await readApplicationsDB();
  const appIndex = applications.findIndex(app => app.id === id);

  if (appIndex === -1) {
    return res.status(404).json({
      success: false,
      message: '申请不存在'
    });
  }

  const app = applications[appIndex];

  // 权限检查
  if (req.user.role === 'dsl_admin' && app.serviceId !== '13') {
    return res.status(403).json({
      success: false,
      message: '无权操作此申请'
    });
  }

  if (req.user.role === 'study_admin' && app.serviceId !== '9') {
    return res.status(403).json({
      success: false,
      message: '无权操作此申请'
    });
  }

  applications[appIndex] = {
    ...app,
    status: status || app.status,
    remark: remark !== undefined ? remark : app.remark,
    updateTime: new Date().toISOString()
  };

  await writeApplicationsDB(applications);

  logger.info('Application updated', {
    applicationId: id,
    status,
    adminId: req.user.id
  });

  res.json({
    success: true,
    data: applications[appIndex]
  });
}));

/**
 * 导出申请为 Excel
 */
router.get('/applications/export', asyncHandler(async (req, res) => {
  let applications = await readApplicationsDB();

  // 根据角色过滤
  if (req.user.role === 'dsl_admin') {
    applications = applications.filter(app => app.serviceId === '13');
  } else if (req.user.role === 'study_admin') {
    applications = applications.filter(app => app.serviceId === '9');
  }

  // 准备导出数据
  const exportData = applications.map(app => ({
    '申请ID': app.id,
    '服务名称': app.serviceName,
    '联系人': app.contactName,
    '联系电话': app.contactPhone,
    '状态': app.status,
    '申请时间': app.createTime,
    '备注': app.remark || ''
  }));

  const workbook = XLSX.utils.book_new();
  const worksheet = XLSX.utils.json_to_sheet(exportData);
  XLSX.utils.book_append_sheet(workbook, worksheet, '申请列表');

  const buffer = XLSX.write(workbook, { type: 'buffer', bookType: 'xlsx' });

  res.setHeader('Content-Disposition', `attachment; filename=applications_${Date.now()}.xlsx`);
  res.setHeader('Content-Type', 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet');

  logger.info('Applications exported', {
    count: applications.length,
    adminId: req.user.id
  });

  res.send(buffer);
}));

/**
 * 获取用户列表
 */
router.get('/users', roleRequired(['admin']), asyncHandler(async (req, res) => {
  const users = await readUsersDB();

  // 清理敏感信息
  const safeUsers = users.map(u => {
    const { password, passwordHash, refreshToken, refreshTokenExpiresAt, wxSessionKey, ...safe } = u;
    return safe;
  });

  res.json({
    success: true,
    data: safeUsers
  });
}));

/**
 * 更新用户角色
 */
router.post('/users/:id/role', sanitizeBody, requireFields('role'), roleRequired(['admin']), asyncHandler(async (req, res) => {
  const { id } = req.params;
  const { role } = req.body;

  const allowedRoles = ['user', 'admin', 'dsl_admin', 'study_admin'];

  if (!allowedRoles.includes(role)) {
    return res.status(400).json({
      success: false,
      message: `无效的角色。允许的角色: ${allowedRoles.join(', ')}`
    });
  }

  const users = await readUsersDB();
  const userIndex = users.findIndex(u => u.id === id);

  if (userIndex === -1) {
    return res.status(404).json({
      success: false,
      message: '用户不存在'
    });
  }

  // 防止修改超级管理员角色
  if (users[userIndex].phone === config.admin.superAdminPhone && role !== 'admin') {
    return res.status(403).json({
      success: false,
      message: '无法修改超级管理员角色'
    });
  }

  const oldRole = users[userIndex].role;
  users[userIndex].role = role;

  await writeUsersDB(users);

  logger.info('User role updated', {
    userId: id,
    oldRole,
    newRole: role,
    adminId: req.user.id
  });

  res.json({
    success: true,
    message: '角色更新成功'
  });
}));

/**
 * 获取服务配置
 */
router.get('/services/config', asyncHandler(async (req, res) => {
  const config = await readServicesConfig();

  // 研学管理员只能看到服务9
  if (req.user.role === 'study_admin') {
    return res.json({
      success: true,
      data: { '9': config['9'] }
    });
  }

  // DSL管理员只能看到服务13
  if (req.user.role === 'dsl_admin') {
    return res.json({
      success: true,
      data: { '13': config['13'] }
    });
  }

  res.json({
    success: true,
    data: config
  });
}));

/**
 * 更新服务配置
 */
router.post('/services/config', sanitizeBody, asyncHandler(async (req, res) => {
  const servicesConfig = await readServicesConfig();
  const newConfig = req.body;

  // 研学管理员只能修改服务9
  if (req.user.role === 'study_admin') {
    if (newConfig['9']) {
      servicesConfig['9'] = { ...servicesConfig['9'], ...newConfig['9'] };
    } else {
      return res.status(403).json({
        success: false,
        message: '无权修改此服务配置'
      });
    }
  }
  // DSL管理员只能修改服务13
  else if (req.user.role === 'dsl_admin') {
    if (newConfig['13']) {
      servicesConfig['13'] = { ...servicesConfig['13'], ...newConfig['13'] };
    } else {
      return res.status(403).json({
        success: false,
        message: '无权修改此服务配置'
      });
    }
  }
  // 超级管理员可以修改所有配置
  else if (req.user.role === 'admin') {
    Object.assign(servicesConfig, newConfig);
  } else {
    return res.status(403).json({
      success: false,
      message: '无权修改服务配置'
    });
  }

  await writeServicesConfig(servicesConfig);

  logger.info('Services config updated', {
    adminId: req.user.id,
    roles: req.user.role
  });

  res.json({
    success: true,
    message: '配置更新成功'
  });
}));

/**
 * 获取研学展示数据
 */
router.get('/study/showcase', asyncHandler(async (req, res) => {
  const config = await readServicesConfig();
  const showcase = config['9']?.studyShowcase || [];

  res.json({
    success: true,
    data: showcase
  });
}));

/**
 * 更新研学展示数据
 */
router.post('/study/showcase', sanitizeBody, asyncHandler(async (req, res) => {
  const { showcase } = req.body;

  if (!Array.isArray(showcase)) {
    return res.status(400).json({
      success: false,
      message: '展示数据必须是数组'
    });
  }

  const servicesConfig = await readServicesConfig();

  if (!servicesConfig['9']) {
    servicesConfig['9'] = {};
  }

  servicesConfig['9'].studyShowcase = showcase;

  await writeServicesConfig(servicesConfig);

  logger.info('Study showcase updated', {
    count: showcase.length,
    adminId: req.user.id
  });

  res.json({
    success: true,
    message: '研学展示更新成功'
  });
}));

/**
 * 获取所有评价（管理员）
 */
router.get('/reviews', asyncHandler(async (req, res) => {
  const { section, status, page = 1, limit = 20 } = req.query;

  let reviews = await readReviewsDB();

  if (section) {
    reviews = reviews.filter(r => r.section === section);
  }
  if (status) {
    reviews = reviews.filter(r => r.status === status);
  }

  reviews.sort((a, b) => new Date(b.createTime) - new Date(a.createTime));

  const total = reviews.length;
  const start = (parseInt(page) - 1) * parseInt(limit);
  const paginated = reviews.slice(start, start + parseInt(limit));

  res.json({
    success: true,
    data: paginated,
    pagination: {
      page: parseInt(page),
      limit: parseInt(limit),
      total,
      totalPages: Math.ceil(total / parseInt(limit))
    }
  });
}));

/**
 * 审核评价（通过/拒绝）
 */
router.post('/reviews/:id', sanitizeBody, asyncHandler(async (req, res) => {
  const { id } = req.params;
  const { status } = req.body;

  if (!status || !['approved', 'rejected'].includes(status)) {
    return res.status(400).json({
      success: false,
      message: '状态必须为 approved 或 rejected'
    });
  }

  const reviews = await readReviewsDB();
  const index = reviews.findIndex(r => r.id === id);

  if (index === -1) {
    return res.status(404).json({
      success: false,
      message: '评价不存在'
    });
  }

  reviews[index] = {
    ...reviews[index],
    status,
    reviewTime: new Date().toISOString()
  };

  await writeReviewsDB(reviews);

  logger.info('Review updated', {
    reviewId: id,
    status,
    adminId: req.user.id
  });

  res.json({
    success: true,
    data: reviews[index]
  });
}));

/**
 * 删除评价
 */
router.delete('/reviews/:id', asyncHandler(async (req, res) => {
  const { id } = req.params;

  const reviews = await readReviewsDB();
  const index = reviews.findIndex(r => r.id === id);

  if (index === -1) {
    return res.status(404).json({
      success: false,
      message: '评价不存在'
    });
  }

  reviews.splice(index, 1);
  await writeReviewsDB(reviews);

  logger.info('Review deleted', {
    reviewId: id,
    adminId: req.user.id
  });

  res.json({
    success: true,
    message: '评价已删除'
  });
}));

module.exports = router;
