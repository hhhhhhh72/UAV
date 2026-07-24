/**
 * 医疗配送模块路由
 * 包含：起降场管理、寄件人认证、常用联系人、配送订单、评价
 */
const express = require('express');
const { randomUUID } = require('crypto');
const { logger } = require('../logger');
const {
  readMedicalOrdersDB, writeMedicalOrdersDB,
  readMedicalCertificationsDB, writeMedicalCertificationsDB,
  readMedicalPadsDB, writeMedicalPadsDB,
  readMedicalContactsDB, writeMedicalContactsDB,
  readMedicalRatingsDB, writeMedicalRatingsDB,
  readMedicalSmsLogsDB, writeMedicalSmsLogsDB
} = require('../storage');
const { authRequired, roleRequired } = require('../middleware/auth');
const { asyncHandler } = require('../middleware/error');

const router = express.Router();

// ==================== 工具函数 ====================

/**
 * 生成订单号: MD + 年月日 + 4位自增序号
 */
async function generateOrderNo() {
  const now = new Date();
  const dateStr = now.getFullYear().toString() +
    String(now.getMonth() + 1).padStart(2, '0') +
    String(now.getDate()).padStart(2, '0');
  const prefix = 'MD' + dateStr;

  const orders = await readMedicalOrdersDB();
  const todayOrders = orders.filter(o => o.order_no && o.order_no.startsWith(prefix));
  const seq = todayOrders.length + 1;
  return prefix + String(seq).padStart(4, '0');
}

/**
 * 预估时间计算
 */
function calculateEstimate(distanceKm, urgency) {
  // 预估时间(分钟)
  const responseTime = { normal: 30, urgent: 15, critical: 5 };
  const prepTime = { normal: 20, urgent: 10, critical: 5 };
  const flySpeed = 60; // km/h
  const flyTime = (distanceKm / flySpeed) * 60;
  const totalMinutes = (responseTime[urgency] || 30) + (prepTime[urgency] || 20) + flyTime;

  return {
    estimatedMinutes: Math.round(totalMinutes)
  };
}

/**
 * 短信通知 - 一期日志模拟
 */
async function sendSmsNotification(phone, templateKey, params, orderId) {
  const templates = {
    order_created: `【低空综合服务平台】您的医疗配送订单 ${params.orderNo || ''} 已提交，等待接单。`,
    order_accepted: `【低空综合服务平台】您的医疗配送订单 ${params.orderNo || ''} 已被接单，正在安排配送。`,
    order_delivering: `【低空综合服务平台】您的医疗配送订单 ${params.orderNo || ''} 正在配送中。`,
    order_delivered: `【低空综合服务平台】您的医疗配送订单 ${params.orderNo || ''} 已送达 ${params.padName || ''}，请尽快前往取货。`,
    order_completed: `【低空综合服务平台】您的医疗配送订单 ${params.orderNo || ''} 已完成。`,
    order_cancelled: `【低空综合服务平台】您的医疗配送订单 ${params.orderNo || ''} 已取消。`,
    order_exception: `【低空综合服务平台】您的医疗配送订单 ${params.orderNo || ''} 出现异常，请关注后续处理。`,
    certification_result: `【低空综合服务平台】您的寄件人认证申请已${params.result || '处理'}。`
  };

  const content = templates[templateKey] || `【低空综合服务平台】通知: ${templateKey}`;

  // 一期: 日志模拟发送
  logger.info('[SMS模拟]', { phone, templateKey, content });

  // 记录发送日志
  const logs = await readMedicalSmsLogsDB();
  logs.push({
    id: randomUUID(),
    order_id: orderId || null,
    phone,
    template_key: templateKey,
    content,
    status: 'simulated',
    provider_msg: '一期日志模拟',
    created_at: new Date().toISOString()
  });
  await writeMedicalSmsLogsDB(logs);
}

// ==================== 起降场管理 ====================

// 获取已启用起降场列表（C端，无需认证）
router.get('/pads', asyncHandler(async (req, res) => {
  const pads = await readMedicalPadsDB();
  const enabledPads = pads.filter(p => p.enabled && !p.deleted);
  res.json({ success: true, data: enabledPads });
}));

// 获取全部起降场（管理端）
router.get('/pads/all', authRequired, roleRequired(['admin', 'dsl_admin']), asyncHandler(async (req, res) => {
  const pads = await readMedicalPadsDB();
  const activePads = pads.filter(p => !p.deleted);
  res.json({ success: true, data: activePads });
}));

// 新增起降场
router.post('/pads', authRequired, roleRequired(['admin', 'dsl_admin']), asyncHandler(async (req, res) => {
  const { name, address, latitude, longitude, contact_name, contact_phone, operating_hours, max_weight } = req.body;

  if (!name || !latitude || !longitude) {
    return res.status(400).json({ success: false, message: '名称和坐标为必填项' });
  }

  const pads = await readMedicalPadsDB();
  const newPad = {
    id: randomUUID(),
    name,
    address: address || '',
    latitude: parseFloat(latitude),
    longitude: parseFloat(longitude),
    contact_name: contact_name || '',
    contact_phone: contact_phone || '',
    operating_hours: operating_hours || '08:00-20:00',
    max_weight: parseFloat(max_weight) || 5.0,
    enabled: true,
    deleted: false,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString()
  };

  pads.push(newPad);
  await writeMedicalPadsDB(pads);

  logger.info('起降场已创建', { padId: newPad.id, name });
  res.json({ success: true, data: newPad });
}));

// 编辑起降场
router.put('/pads/:id', authRequired, roleRequired(['admin', 'dsl_admin']), asyncHandler(async (req, res) => {
  const { id } = req.params;
  const pads = await readMedicalPadsDB();
  const index = pads.findIndex(p => p.id === id && !p.deleted);

  if (index === -1) {
    return res.status(404).json({ success: false, message: '起降场不存在' });
  }

  const allowedFields = ['name', 'address', 'latitude', 'longitude', 'contact_name', 'contact_phone', 'operating_hours', 'max_weight', 'enabled'];
  for (const field of allowedFields) {
    if (req.body[field] !== undefined) {
      if (['latitude', 'longitude', 'max_weight'].includes(field)) {
        pads[index][field] = parseFloat(req.body[field]);
      } else if (field === 'enabled') {
        pads[index][field] = Boolean(req.body[field]);
      } else {
        pads[index][field] = req.body[field];
      }
    }
  }
  pads[index].updated_at = new Date().toISOString();

  await writeMedicalPadsDB(pads);
  res.json({ success: true, data: pads[index] });
}));

// 删除起降场（软删除）
router.delete('/pads/:id', authRequired, roleRequired(['admin', 'dsl_admin']), asyncHandler(async (req, res) => {
  const { id } = req.params;
  const pads = await readMedicalPadsDB();
  const index = pads.findIndex(p => p.id === id && !p.deleted);

  if (index === -1) {
    return res.status(404).json({ success: false, message: '起降场不存在' });
  }

  // 检查是否有关联订单
  const orders = await readMedicalOrdersDB();
  const hasOrders = orders.some(o =>
    (o.route.departure_pad_id === id || o.route.arrival_pad_id === id) &&
    !['cancelled', 'completed'].includes(o.status)
  );

  if (hasOrders) {
    return res.status(400).json({ success: false, message: '该起降场有进行中的订单，无法删除' });
  }

  pads[index].deleted = true;
  pads[index].updated_at = new Date().toISOString();
  await writeMedicalPadsDB(pads);

  res.json({ success: true, message: '删除成功' });
}));

// ==================== 寄件人认证 ====================

// 查询当前用户认证状态
router.get('/certification/status', authRequired, asyncHandler(async (req, res) => {
  const certs = await readMedicalCertificationsDB();
  const myCert = certs.find(c => c.user_id === req.user.id);

  if (!myCert) {
    return res.json({ success: true, data: { status: 'none' } });
  }

  // 隐藏敏感信息（历史数据兼容）
  const safeData = { ...myCert };
  delete safeData.id_card;
  delete safeData.id_card_images;
  res.json({ success: true, data: safeData });
}));

// 提交认证申请
router.post('/certification/apply', authRequired, asyncHandler(async (req, res) => {
  const { real_name, phone, org_type, org_name, org_address, position } = req.body;

  if (!real_name || !phone || !org_type || !org_name) {
    return res.status(400).json({ success: false, message: '请填写完整的认证信息' });
  }

  const certs = await readMedicalCertificationsDB();
  const existing = certs.find(c => c.user_id === req.user.id);

  if (existing && existing.status === 'approved') {
    return res.status(400).json({ success: false, message: '您已通过认证' });
  }
  if (existing && existing.status === 'pending') {
    return res.status(400).json({ success: false, message: '您的认证正在审核中' });
  }

  const newCert = {
    id: randomUUID(),
    user_id: req.user.id,
    real_name,
    phone,
    org_type,
    org_name,
    org_address: org_address || '',
    position: position || '',
    status: 'pending',
    auth_agreement: true,
    auth_signed_at: new Date().toISOString(),
    review: { reviewed_by: null, reviewed_at: null, reject_reason: null },
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString()
  };

  // 如果有之前驳回的记录，替换
  if (existing) {
    const index = certs.findIndex(c => c.user_id === req.user.id);
    certs[index] = newCert;
  } else {
    certs.push(newCert);
  }

  await writeMedicalCertificationsDB(certs);
  res.json({ success: true, message: '认证申请已提交，请等待审核' });
}));

// 重新提交认证（驳回后）
router.post('/certification/resubmit', authRequired, asyncHandler(async (req, res) => {
  const certs = await readMedicalCertificationsDB();
  const existing = certs.find(c => c.user_id === req.user.id);

  if (!existing || existing.status !== 'rejected') {
    return res.status(400).json({ success: false, message: '无需重新提交' });
  }

  const { real_name, phone, org_type, org_name, org_address, position } = req.body;

  if (!real_name || !phone || !org_type || !org_name) {
    return res.status(400).json({ success: false, message: '请填写完整的认证信息' });
  }

  const index = certs.findIndex(c => c.user_id === req.user.id);
  certs[index] = {
    ...certs[index],
    real_name, phone,
    org_type, org_name,
    org_address: org_address || '',
    position: position || '',
    status: 'pending',
    review: { reviewed_by: null, reviewed_at: null, reject_reason: null },
    updated_at: new Date().toISOString()
  };

  await writeMedicalCertificationsDB(certs);
  res.json({ success: true, message: '认证已重新提交' });
}));

// 管理端：获取认证列表
router.get('/certifications', authRequired, roleRequired(['admin', 'dsl_admin']), asyncHandler(async (req, res) => {
  const { status, page = 1, limit = 20 } = req.query;
  let certs = await readMedicalCertificationsDB();

  if (status) {
    certs = certs.filter(c => c.status === status);
  }

  certs.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));

  const total = certs.length;
  const start = (parseInt(page) - 1) * parseInt(limit);
  const paginated = certs.slice(start, start + parseInt(limit));

  // 隐藏身份证号
  const safeData = paginated.map(c => ({ ...c, id_card: c.id_card ? '***' : '' }));

  res.json({ success: true, data: safeData, total, page: parseInt(page), limit: parseInt(limit) });
}));

// 管理端：通过认证
router.post('/certifications/:id/approve', authRequired, roleRequired(['admin', 'dsl_admin']), asyncHandler(async (req, res) => {
  const { id } = req.params;
  const certs = await readMedicalCertificationsDB();
  const index = certs.findIndex(c => c.id === id);

  if (index === -1) {
    return res.status(404).json({ success: false, message: '认证记录不存在' });
  }
  if (certs[index].status !== 'pending') {
    return res.status(400).json({ success: false, message: '该认证不在待审核状态' });
  }

  certs[index].status = 'approved';
  certs[index].review = {
    reviewed_by: req.user.id,
    reviewed_at: new Date().toISOString(),
    reject_reason: null
  };
  certs[index].updated_at = new Date().toISOString();

  await writeMedicalCertificationsDB(certs);

  // 发送通知
  await sendSmsNotification(certs[index].phone, 'certification_result', { result: '通过' });

  res.json({ success: true, message: '认证已通过' });
}));

// 管理端：驳回认证
router.post('/certifications/:id/reject', authRequired, roleRequired(['admin', 'dsl_admin']), asyncHandler(async (req, res) => {
  const { id } = req.params;
  const { reason } = req.body;

  if (!reason) {
    return res.status(400).json({ success: false, message: '请填写驳回原因' });
  }

  const certs = await readMedicalCertificationsDB();
  const index = certs.findIndex(c => c.id === id);

  if (index === -1) {
    return res.status(404).json({ success: false, message: '认证记录不存在' });
  }
  if (certs[index].status !== 'pending') {
    return res.status(400).json({ success: false, message: '该认证不在待审核状态' });
  }

  certs[index].status = 'rejected';
  certs[index].review = {
    reviewed_by: req.user.id,
    reviewed_at: new Date().toISOString(),
    reject_reason: reason
  };
  certs[index].updated_at = new Date().toISOString();

  await writeMedicalCertificationsDB(certs);

  // 发送通知
  await sendSmsNotification(certs[index].phone, 'certification_result', { result: '驳回，原因: ' + reason });

  res.json({ success: true, message: '已驳回' });
}));

// ==================== 常用联系人 ====================

// 获取联系人列表
router.get('/contacts', authRequired, asyncHandler(async (req, res) => {
  const contacts = await readMedicalContactsDB();
  const myContacts = contacts.filter(c => c.user_id === req.user.id);
  res.json({ success: true, data: myContacts });
}));

// 新增联系人
router.post('/contacts', authRequired, asyncHandler(async (req, res) => {
  const { name, phone, org_name, label, is_default } = req.body;

  if (!name || !phone) {
    return res.status(400).json({ success: false, message: '姓名和电话为必填项' });
  }

  const contacts = await readMedicalContactsDB();
  const myContacts = contacts.filter(c => c.user_id === req.user.id);

  if (myContacts.length >= 20) {
    return res.status(400).json({ success: false, message: '常用联系人最多20个' });
  }

  const newContact = {
    id: randomUUID(),
    user_id: req.user.id,
    name,
    phone,
    org_name: org_name || '',
    label: label || '',
    is_default: is_default || false,
    created_at: new Date().toISOString()
  };

  contacts.push(newContact);
  await writeMedicalContactsDB(contacts);
  res.json({ success: true, data: newContact });
}));

// 编辑联系人
router.put('/contacts/:id', authRequired, asyncHandler(async (req, res) => {
  const { id } = req.params;
  const contacts = await readMedicalContactsDB();
  const index = contacts.findIndex(c => c.id === id && c.user_id === req.user.id);

  if (index === -1) {
    return res.status(404).json({ success: false, message: '联系人不存在' });
  }

  const { name, phone, org_name, label, is_default } = req.body;
  if (name !== undefined) contacts[index].name = name;
  if (phone !== undefined) contacts[index].phone = phone;
  if (org_name !== undefined) contacts[index].org_name = org_name;
  if (label !== undefined) contacts[index].label = label;
  if (is_default !== undefined) contacts[index].is_default = is_default;

  await writeMedicalContactsDB(contacts);
  res.json({ success: true, data: contacts[index] });
}));

// 删除联系人
router.delete('/contacts/:id', authRequired, asyncHandler(async (req, res) => {
  const { id } = req.params;
  const contacts = await readMedicalContactsDB();
  const index = contacts.findIndex(c => c.id === id && c.user_id === req.user.id);

  if (index === -1) {
    return res.status(404).json({ success: false, message: '联系人不存在' });
  }

  contacts.splice(index, 1);
  await writeMedicalContactsDB(contacts);
  res.json({ success: true, message: '删除成功' });
}));

// ==================== 配送订单 ====================

// 预估费用与时间
router.get('/orders/estimate', asyncHandler(async (req, res) => {
  const { departure_pad_id, arrival_pad_id, urgency, temp_requirements } = req.query;

  if (!departure_pad_id || !arrival_pad_id) {
    return res.status(400).json({ success: false, message: '请选择起降场' });
  }

  const pads = await readMedicalPadsDB();
  const departure = pads.find(p => p.id === departure_pad_id && p.enabled && !p.deleted);
  const arrival = pads.find(p => p.id === arrival_pad_id && p.enabled && !p.deleted);

  if (!departure || !arrival) {
    return res.status(400).json({ success: false, message: '起降场无效' });
  }

  // 计算直线距离(km)
  const R = 6371;
  const dLat = (arrival.latitude - departure.latitude) * Math.PI / 180;
  const dLon = (arrival.longitude - departure.longitude) * Math.PI / 180;
  const a = Math.sin(dLat / 2) * Math.sin(dLat / 2) +
    Math.cos(departure.latitude * Math.PI / 180) * Math.cos(arrival.latitude * Math.PI / 180) *
    Math.sin(dLon / 2) * Math.sin(dLon / 2);
  const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
  const distanceKm = R * c;

  const tempReqs = temp_requirements ? temp_requirements.split(',') : [];
  const estimate = calculateEstimate(distanceKm, urgency || 'normal');

  res.json({
    success: true,
    data: {
      distance_km: Math.round(distanceKm * 10) / 10,
      estimated_minutes: estimate.estimatedMinutes
    }
  });
}));

// 创建配送订单
router.post('/orders', authRequired, asyncHandler(async (req, res) => {
  // 检查寄件人认证状态
  const certs = await readMedicalCertificationsDB();
  const myCert = certs.find(c => c.user_id === req.user.id && c.status === 'approved');
  if (!myCert) {
    return res.status(403).json({ success: false, message: '您尚未通过寄件人认证，无法下单' });
  }

  const {
    sender_name, sender_phone, sender_org,
    receiver_name, receiver_phone, receiver_org,
    departure_pad_id, arrival_pad_id,
    item_type, item_weight, temp_requirements, item_images, item_description,
    urgency
  } = req.body;

  // 校验必填
  if (!sender_name || !sender_phone || !sender_org ||
    !receiver_name || !receiver_phone || !receiver_org ||
    !departure_pad_id || !arrival_pad_id ||
    !item_type || !item_weight || !urgency) {
    return res.status(400).json({ success: false, message: '请填写完整的订单信息' });
  }

  // 校验收货人也已通过认证（通过手机号匹配）
  const receiverCert = certs.find(c => c.phone === receiver_phone && c.status === 'approved');
  if (!receiverCert) {
    return res.status(400).json({ success: false, message: '收货人尚未完成身份认证，请通知对方先在平台完成认证后再下单' });
  }

  // 校验图片
  if (!item_images || !Array.isArray(item_images) || item_images.length < 1 || item_images.length > 3) {
    return res.status(400).json({ success: false, message: '请上传1-3张物品照片' });
  }

  // 校验起降场
  const pads = await readMedicalPadsDB();
  const departure = pads.find(p => p.id === departure_pad_id && p.enabled && !p.deleted);
  const arrival = pads.find(p => p.id === arrival_pad_id && p.enabled && !p.deleted);
  if (!departure || !arrival) {
    return res.status(400).json({ success: false, message: '起降场无效或已停用' });
  }
  if (departure_pad_id === arrival_pad_id) {
    return res.status(400).json({ success: false, message: '出发与到达不能为同一起降场' });
  }

  // 校验重量
  const weight = parseFloat(item_weight);
  if (weight < 0.1 || weight > 10) {
    return res.status(400).json({ success: false, message: '物品重量需在0.1-10kg之间' });
  }

  // 计算距离和预估
  const R = 6371;
  const dLat = (arrival.latitude - departure.latitude) * Math.PI / 180;
  const dLon = (arrival.longitude - departure.longitude) * Math.PI / 180;
  const a = Math.sin(dLat / 2) * Math.sin(dLat / 2) +
    Math.cos(departure.latitude * Math.PI / 180) * Math.cos(arrival.latitude * Math.PI / 180) *
    Math.sin(dLon / 2) * Math.sin(dLon / 2);
  const c2 = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
  const distanceKm = R * c2;

  const tempReqs = temp_requirements || [];
  const estimate = calculateEstimate(distanceKm, urgency);

  const orderNo = await generateOrderNo();
  const now = new Date().toISOString();
  const estimatedDeliveryTime = new Date(Date.now() + estimate.estimatedMinutes * 60000).toISOString();

  // 物品类型标签映射
  const typeLabels = { blood: '血液制品', sample: '检验样本', medicine: '药品制剂', equipment: '医疗器械', other: '其他' };
  const urgencyLabels = { normal: '普通', urgent: '加急', critical: '特急' };
  const tempLabels = { normal: '常温', refrigerated: '冷藏(2-8℃)', frozen: '冷冻(-20℃)', lightproof: '避光' };

  const newOrder = {
    id: randomUUID(),
    order_no: orderNo,
    sender: {
      user_id: req.user.id,
      name: sender_name,
      phone: sender_phone,
      org: sender_org,
      certification_id: myCert.id
    },
    receiver: {
      user_id: receiverCert.user_id || null,
      name: receiver_name,
      phone: receiver_phone,
      org: receiver_org,
      certification_id: receiverCert.id
    },
    route: {
      departure_pad_id,
      departure_name: departure.name,
      arrival_pad_id,
      arrival_name: arrival.name,
      distance_km: Math.round(distanceKm * 10) / 10
    },
    item: {
      type: item_type,
      type_label: typeLabels[item_type] || item_type,
      weight,
      temp_requirement: tempReqs,
      temp_labels: tempReqs.map(t => tempLabels[t] || t),
      images: item_images,
      description: item_description || ''
    },
    urgency,
    urgency_label: urgencyLabels[urgency] || urgency,
    estimated_delivery_time: estimatedDeliveryTime,
    status: 'pending',
    status_label: '待接单',
    timeline: [
      { status: 'pending', label: '待接单', time: now, operator: 'system' }
    ],
    operator: {
      accepted_by: null, accepted_at: null,
      pickup_at: null,
      delivered_by: null, delivered_at: null,
      completed_by: null, completed_at: null
    },
    cancel_info: null,
    exception_info: null,
    rating: null,
    remark: '',
    timeout_warned: false,
    created_at: now,
    updated_at: now
  };

  const orders = await readMedicalOrdersDB();
  orders.push(newOrder);
  await writeMedicalOrdersDB(orders);

  // 发送通知
  await sendSmsNotification(sender_phone, 'order_created', { orderNo }, newOrder.id);
  await sendSmsNotification(receiver_phone, 'order_created', { orderNo }, newOrder.id);

  logger.info('医疗配送订单已创建', { orderId: newOrder.id, orderNo });
  res.json({ success: true, data: { id: newOrder.id, order_no: orderNo } });
}));

// C端：我的订单列表
router.get('/orders/my', authRequired, asyncHandler(async (req, res) => {
  const { status, page = 1, limit = 20 } = req.query;
  let orders = await readMedicalOrdersDB();

  orders = orders.filter(o => o.sender.user_id === req.user.id);

  if (status) {
    orders = orders.filter(o => o.status === status);
  }

  orders.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));

  const total = orders.length;
  const start = (parseInt(page) - 1) * parseInt(limit);
  const paginated = orders.slice(start, start + parseInt(limit));

  res.json({ success: true, data: paginated, total, page: parseInt(page), limit: parseInt(limit) });
}));

// C端：收件人订单列表（寄给我的）
router.get('/orders/received', authRequired, asyncHandler(async (req, res) => {
  const { status, page = 1, limit = 20 } = req.query;
  let orders = await readMedicalOrdersDB();

  // 通过 receiver.user_id 或者手机号匹配（兼容旧数据）
  const certs = await readMedicalCertificationsDB();
  const myCert = certs.find(c => c.user_id === req.user.id && c.status === 'approved');
  const myPhone = myCert ? myCert.phone : null;

  orders = orders.filter(o => {
    if (o.receiver?.user_id === req.user.id) return true;
    if (myPhone && o.receiver?.phone === myPhone) return true;
    return false;
  });

  if (status) {
    orders = orders.filter(o => o.status === status);
  }

  orders.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));

  const total = orders.length;
  const start = (parseInt(page) - 1) * parseInt(limit);
  const paginated = orders.slice(start, start + parseInt(limit));

  res.json({ success: true, data: paginated, total, page: parseInt(page), limit: parseInt(limit) });
}));

// C端：收件人签收确认
router.post('/orders/:id/confirm-receipt', authRequired, asyncHandler(async (req, res) => {
  const { id } = req.params;
  const orders = await readMedicalOrdersDB();
  const index = orders.findIndex(o => o.id === id);

  if (index === -1) {
    return res.status(404).json({ success: false, message: '订单不存在' });
  }

  const order = orders[index];

  // 校验是否是收件人
  const certs = await readMedicalCertificationsDB();
  const myCert = certs.find(c => c.user_id === req.user.id && c.status === 'approved');
  const myPhone = myCert ? myCert.phone : null;
  const isReceiver = order.receiver?.user_id === req.user.id || (myPhone && order.receiver?.phone === myPhone);

  if (!isReceiver) {
    return res.status(403).json({ success: false, message: '只有收件人才能确认签收' });
  }

  // 只有“已送达”状态可以签收
  if (order.status !== 'delivered') {
    return res.status(400).json({ success: false, message: '订单当前状态不可签收' });
  }

  const now = new Date().toISOString();
  orders[index].status = 'completed';
  orders[index].status_label = '已签收';
  orders[index].receipt_info = {
    confirmed_by: req.user.id,
    confirmed_at: now,
    confirm_method: 'manual'
  };
  orders[index].operator.completed_at = now;
  orders[index].timeline.push({ status: 'completed', label: '收件人已签收', time: now, operator: req.user.id });
  orders[index].updated_at = now;

  await writeMedicalOrdersDB(orders);

  // 通知寄件人
  await sendSmsNotification(order.sender.phone, 'order_completed', { orderNo: order.order_no }, id);

  logger.info('收件人确认签收', { orderId: id, receiverId: req.user.id });
  res.json({ success: true, message: '签收成功' });
}));

// 订单详情
router.get('/orders/:id', authRequired, asyncHandler(async (req, res) => {
  const { id } = req.params;
  const orders = await readMedicalOrdersDB();
  const order = orders.find(o => o.id === id);

  if (!order) {
    return res.status(404).json({ success: false, message: '订单不存在' });
  }

  // 普通用户只能看自己的订单（寄件人或收件人）
  if (req.user.role === 'user' && order.sender.user_id !== req.user.id && order.receiver?.user_id !== req.user.id) {
    return res.status(403).json({ success: false, message: '无权查看此订单' });
  }

  res.json({ success: true, data: order });
}));

// 取消订单
router.post('/orders/:id/cancel', authRequired, asyncHandler(async (req, res) => {
  const { id } = req.params;
  const { reason_type, reason_detail } = req.body;
  const isAdmin = ['admin', 'dsl_admin'].includes(req.user.role);

  const orders = await readMedicalOrdersDB();
  const index = orders.findIndex(o => o.id === id);

  if (index === -1) {
    return res.status(404).json({ success: false, message: '订单不存在' });
  }

  const order = orders[index];

  // 权限检查
  if (!isAdmin && order.sender.user_id !== req.user.id) {
    return res.status(403).json({ success: false, message: '无权操作此订单' });
  }

  // 取消规则
  if (!isAdmin) {
    if (order.status === 'pending') {
      // 用户可自由取消
    } else if (order.status === 'accepted') {
      // 接单后10分钟内可取消
      const acceptedAt = new Date(order.operator.accepted_at).getTime();
      if (Date.now() - acceptedAt > 10 * 60 * 1000) {
        return res.status(400).json({ success: false, message: '接单超过10分钟，请联系管理员取消' });
      }
    } else {
      return res.status(400).json({ success: false, message: '当前状态不允许取消' });
    }
  } else {
    // 管理端：配送中仅异常可取消
    if (order.status === 'delivering' && order.status !== 'exception') {
      // 管理端可以强制取消配送中的
    }
  }

  const reasonLabels = {
    not_needed: '不需要了',
    info_error: '信息填写错误，重新下单',
    wait_too_long: '等待时间过长',
    other_delivery: '找到其他配送方式',
    other: '其他原因'
  };

  const now = new Date().toISOString();
  orders[index].status = 'cancelled';
  orders[index].status_label = '已取消';
  orders[index].cancel_info = {
    cancelled_by: isAdmin ? 'admin' : 'user',
    cancelled_at: now,
    reason_type: reason_type || 'other',
    reason_label: reasonLabels[reason_type] || '其他原因',
    reason_detail: reason_detail || '',
    cancel_stage: order.status
  };
  orders[index].timeline.push({ status: 'cancelled', label: '已取消', time: now, operator: req.user.id });
  orders[index].updated_at = now;

  await writeMedicalOrdersDB(orders);

  // 通知双方
  await sendSmsNotification(order.sender.phone, 'order_cancelled', { orderNo: order.order_no }, order.id);
  await sendSmsNotification(order.receiver.phone, 'order_cancelled', { orderNo: order.order_no }, order.id);

  res.json({ success: true, message: '订单已取消' });
}));

// 再次下单（获取复制数据）
router.post('/orders/:id/reorder', authRequired, asyncHandler(async (req, res) => {
  const { id } = req.params;
  const orders = await readMedicalOrdersDB();
  const order = orders.find(o => o.id === id && o.sender.user_id === req.user.id);

  if (!order) {
    return res.status(404).json({ success: false, message: '订单不存在' });
  }

  // 返回可复制的字段
  res.json({
    success: true,
    data: {
      sender_name: order.sender.name,
      sender_phone: order.sender.phone,
      sender_org: order.sender.org,
      receiver_name: order.receiver.name,
      receiver_phone: order.receiver.phone,
      receiver_org: order.receiver.org,
      departure_pad_id: order.route.departure_pad_id,
      arrival_pad_id: order.route.arrival_pad_id,
      item_type: order.item.type,
      temp_requirements: order.item.temp_requirement
    }
  });
}));

// 管理端：全部订单列表
router.get('/orders', authRequired, roleRequired(['admin', 'dsl_admin']), asyncHandler(async (req, res) => {
  const { status, urgency, page = 1, limit = 20, search, departure_pad_id } = req.query;
  let orders = await readMedicalOrdersDB();

  if (status) orders = orders.filter(o => o.status === status);
  if (urgency) orders = orders.filter(o => o.urgency === urgency);
  if (departure_pad_id) orders = orders.filter(o => o.route.departure_pad_id === departure_pad_id);
  if (search) {
    const s = search.toLowerCase();
    orders = orders.filter(o =>
      o.order_no.toLowerCase().includes(s) ||
      o.sender.name.toLowerCase().includes(s) ||
      o.sender.phone.includes(s) ||
      o.receiver.name.toLowerCase().includes(s)
    );
  }

  orders.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));

  const total = orders.length;
  const start = (parseInt(page) - 1) * parseInt(limit);
  const paginated = orders.slice(start, start + parseInt(limit));

  // 统计待接单数量
  const allOrders = await readMedicalOrdersDB();
  const pendingCount = allOrders.filter(o => o.status === 'pending').length;

  res.json({ success: true, data: paginated, total, page: parseInt(page), limit: parseInt(limit), pendingCount });
}));

// 管理端：接单
router.post('/orders/:id/accept', authRequired, roleRequired(['admin', 'dsl_admin']), asyncHandler(async (req, res) => {
  const { id } = req.params;
  const orders = await readMedicalOrdersDB();
  const index = orders.findIndex(o => o.id === id);

  if (index === -1) return res.status(404).json({ success: false, message: '订单不存在' });
  if (orders[index].status !== 'pending') return res.status(400).json({ success: false, message: '订单不在待接单状态' });

  const now = new Date().toISOString();
  orders[index].status = 'accepted';
  orders[index].status_label = '已接单';
  orders[index].operator.accepted_by = req.user.id;
  orders[index].operator.accepted_at = now;
  orders[index].timeline.push({ status: 'accepted', label: '已接单', time: now, operator: req.user.id });
  orders[index].updated_at = now;

  await writeMedicalOrdersDB(orders);
  await sendSmsNotification(orders[index].sender.phone, 'order_accepted', { orderNo: orders[index].order_no }, id);
  await sendSmsNotification(orders[index].receiver.phone, 'order_accepted', { orderNo: orders[index].order_no }, id);

  res.json({ success: true, message: '已接单' });
}));

// 管理端：标记待取件
router.post('/orders/:id/pickup', authRequired, roleRequired(['admin', 'dsl_admin']), asyncHandler(async (req, res) => {
  const { id } = req.params;
  const orders = await readMedicalOrdersDB();
  const index = orders.findIndex(o => o.id === id);

  if (index === -1) return res.status(404).json({ success: false, message: '订单不存在' });
  if (orders[index].status !== 'accepted') return res.status(400).json({ success: false, message: '订单不在已接单状态' });

  const now = new Date().toISOString();
  orders[index].status = 'pickup';
  orders[index].status_label = '待取件';
  orders[index].operator.pickup_at = now;
  orders[index].timeline.push({ status: 'pickup', label: '待取件', time: now, operator: req.user.id });
  orders[index].updated_at = now;

  await writeMedicalOrdersDB(orders);
  res.json({ success: true, message: '已标记待取件' });
}));

// 管理端：标记配送中
router.post('/orders/:id/deliver', authRequired, roleRequired(['admin', 'dsl_admin']), asyncHandler(async (req, res) => {
  const { id } = req.params;
  const orders = await readMedicalOrdersDB();
  const index = orders.findIndex(o => o.id === id);

  if (index === -1) return res.status(404).json({ success: false, message: '订单不存在' });
  if (!['accepted', 'pickup'].includes(orders[index].status)) return res.status(400).json({ success: false, message: '当前状态不允许此操作' });

  const now = new Date().toISOString();
  orders[index].status = 'delivering';
  orders[index].status_label = '配送中';
  orders[index].timeline.push({ status: 'delivering', label: '配送中', time: now, operator: req.user.id });
  orders[index].updated_at = now;

  await writeMedicalOrdersDB(orders);
  await sendSmsNotification(orders[index].sender.phone, 'order_delivering', { orderNo: orders[index].order_no }, id);
  await sendSmsNotification(orders[index].receiver.phone, 'order_delivering', { orderNo: orders[index].order_no }, id);

  res.json({ success: true, message: '已标记配送中' });
}));

// 管理端：标记已送达
router.post('/orders/:id/delivered', authRequired, roleRequired(['admin', 'dsl_admin']), asyncHandler(async (req, res) => {
  const { id } = req.params;
  const orders = await readMedicalOrdersDB();
  const index = orders.findIndex(o => o.id === id);

  if (index === -1) return res.status(404).json({ success: false, message: '订单不存在' });
  if (orders[index].status !== 'delivering') return res.status(400).json({ success: false, message: '订单不在配送中状态' });

  const now = new Date().toISOString();
  orders[index].status = 'delivered';
  orders[index].status_label = '已送达';
  orders[index].operator.delivered_by = req.user.id;
  orders[index].operator.delivered_at = now;
  orders[index].timeline.push({ status: 'delivered', label: '已送达', time: now, operator: req.user.id });
  orders[index].updated_at = now;

  await writeMedicalOrdersDB(orders);

  const padName = orders[index].route.arrival_name;
  await sendSmsNotification(orders[index].sender.phone, 'order_delivered', { orderNo: orders[index].order_no, padName }, id);
  await sendSmsNotification(orders[index].receiver.phone, 'order_delivered', { orderNo: orders[index].order_no, padName }, id);

  res.json({ success: true, message: '已标记送达' });
}));

// 管理端：标记已完成
router.post('/orders/:id/complete', authRequired, roleRequired(['admin', 'dsl_admin']), asyncHandler(async (req, res) => {
  const { id } = req.params;
  const orders = await readMedicalOrdersDB();
  const index = orders.findIndex(o => o.id === id);

  if (index === -1) return res.status(404).json({ success: false, message: '订单不存在' });
  if (orders[index].status !== 'delivered') return res.status(400).json({ success: false, message: '订单不在已送达状态' });

  const now = new Date().toISOString();
  orders[index].status = 'completed';
  orders[index].status_label = '已完成';
  orders[index].operator.completed_by = req.user.id;
  orders[index].operator.completed_at = now;
  orders[index].timeline.push({ status: 'completed', label: '已完成', time: now, operator: req.user.id });
  orders[index].updated_at = now;

  await writeMedicalOrdersDB(orders);
  await sendSmsNotification(orders[index].sender.phone, 'order_completed', { orderNo: orders[index].order_no }, id);

  res.json({ success: true, message: '订单已完成' });
}));

// 管理端：标记异常
router.post('/orders/:id/exception', authRequired, roleRequired(['admin', 'dsl_admin']), asyncHandler(async (req, res) => {
  const { id } = req.params;
  const { exception_type, description } = req.body;

  if (!exception_type) {
    return res.status(400).json({ success: false, message: '请选择异常类型' });
  }

  const orders = await readMedicalOrdersDB();
  const index = orders.findIndex(o => o.id === id);

  if (index === -1) return res.status(404).json({ success: false, message: '订单不存在' });
  if (['completed', 'cancelled'].includes(orders[index].status)) {
    return res.status(400).json({ success: false, message: '已完成/已取消订单不可标记异常' });
  }

  const exceptionLabels = {
    WEATHER: '天气原因', EQUIPMENT: '设备故障', AIRSPACE: '空域限制',
    ITEM_ISSUE: '物品问题', RECEIVER: '取货人问题', OTHER: '其他'
  };

  const now = new Date().toISOString();
  orders[index].status = 'exception';
  orders[index].status_label = '异常';
  orders[index].exception_info = {
    type: exception_type,
    label: exceptionLabels[exception_type] || exception_type,
    description: description || '',
    marked_by: req.user.id,
    marked_at: now
  };
  orders[index].timeline.push({ status: 'exception', label: '异常: ' + (exceptionLabels[exception_type] || exception_type), time: now, operator: req.user.id });
  orders[index].updated_at = now;

  await writeMedicalOrdersDB(orders);
  await sendSmsNotification(orders[index].sender.phone, 'order_exception', { orderNo: orders[index].order_no }, id);
  await sendSmsNotification(orders[index].receiver.phone, 'order_exception', { orderNo: orders[index].order_no }, id);

  res.json({ success: true, message: '已标记异常' });
}));

// ==================== 评价 ====================

// 提交评价
router.post('/orders/:id/rate', authRequired, asyncHandler(async (req, res) => {
  const { id } = req.params;
  const { rating, tags, comment } = req.body;

  if (!rating || rating < 1 || rating > 5) {
    return res.status(400).json({ success: false, message: '评分需在1-5之间' });
  }

  const orders = await readMedicalOrdersDB();
  const order = orders.find(o => o.id === id && o.sender.user_id === req.user.id);

  if (!order) return res.status(404).json({ success: false, message: '订单不存在' });
  if (order.status !== 'completed') return res.status(400).json({ success: false, message: '仅已完成订单可评价' });
  if (order.rating) return res.status(400).json({ success: false, message: '该订单已评价' });

  const ratings = await readMedicalRatingsDB();
  const newRating = {
    id: randomUUID(),
    order_id: id,
    user_id: req.user.id,
    rating: parseInt(rating),
    tags: tags || [],
    comment: comment || '',
    created_at: new Date().toISOString()
  };
  ratings.push(newRating);
  await writeMedicalRatingsDB(ratings);

  // 更新订单关联
  const orderIndex = orders.findIndex(o => o.id === id);
  orders[orderIndex].rating = newRating.id;
  await writeMedicalOrdersDB(orders);

  res.json({ success: true, data: newRating });
}));

// 获取订单评价
router.get('/orders/:id/rating', authRequired, asyncHandler(async (req, res) => {
  const { id } = req.params;
  const ratings = await readMedicalRatingsDB();
  const rating = ratings.find(r => r.order_id === id);

  if (!rating) {
    return res.json({ success: true, data: null });
  }
  res.json({ success: true, data: rating });
}));

// 管理端：评价统计
router.get('/ratings/stats', authRequired, roleRequired(['admin', 'dsl_admin']), asyncHandler(async (req, res) => {
  const ratings = await readMedicalRatingsDB();

  const total = ratings.length;
  const avgRating = total > 0 ? (ratings.reduce((sum, r) => sum + r.rating, 0) / total).toFixed(1) : 0;
  const distribution = { 5: 0, 4: 0, 3: 0, 2: 0, 1: 0 };
  ratings.forEach(r => { distribution[r.rating] = (distribution[r.rating] || 0) + 1; });

  res.json({
    success: true,
    data: { total, avgRating: parseFloat(avgRating), distribution, recent: ratings.slice(-10).reverse() }
  });
}));

// ==================== 超时检查定时任务 ====================

async function checkOrderTimeouts() {
  try {
    const orders = await readMedicalOrdersDB();
    const now = Date.now();
    let modified = false;

    for (let i = 0; i < orders.length; i++) {
      const order = orders[i];
      if (order.timeout_warned) continue;

      const createdAt = new Date(order.created_at).getTime();
      const acceptedAt = order.operator.accepted_at ? new Date(order.operator.accepted_at).getTime() : null;
      const deliveredAt = order.operator.delivered_at ? new Date(order.operator.delivered_at).getTime() : null;

      // 待接单超时
      if (order.status === 'pending') {
        const timeoutMs = { normal: 30 * 60000, urgent: 15 * 60000, critical: 5 * 60000 };
        const timeout = timeoutMs[order.urgency] || 30 * 60000;
        if (now - createdAt > timeout) {
          orders[i].timeout_warned = true;
          orders[i].timeline.push({ status: 'timeout_warn', label: '待接单超时提醒', time: new Date().toISOString(), operator: 'system' });
          modified = true;
          logger.warn('订单待接单超时', { orderId: order.id, orderNo: order.order_no });
        }
      }

      // 已接单未配送超时 (60分钟)
      if (order.status === 'accepted' && acceptedAt) {
        if (now - acceptedAt > 60 * 60000) {
          orders[i].timeout_warned = true;
          orders[i].timeline.push({ status: 'timeout_warn', label: '已接单未配送超时提醒', time: new Date().toISOString(), operator: 'system' });
          modified = true;
          logger.warn('订单已接单未配送超时', { orderId: order.id, orderNo: order.order_no });
        }
      }

      // 已送达未取件超时 (2小时提醒，4小时标记异常)
      if (order.status === 'delivered' && deliveredAt) {
        if (now - deliveredAt > 4 * 60 * 60000) {
          orders[i].status = 'exception';
          orders[i].status_label = '异常';
          orders[i].exception_info = { type: 'RECEIVER', label: '取货人问题', description: '送达超4小时未取件，系统自动标记异常', marked_by: 'system', marked_at: new Date().toISOString() };
          orders[i].timeline.push({ status: 'exception', label: '超时自动标记异常', time: new Date().toISOString(), operator: 'system' });
          orders[i].timeout_warned = true;
          modified = true;
          logger.warn('订单送达超时自动标记异常', { orderId: order.id, orderNo: order.order_no });
        } else if (now - deliveredAt > 2 * 60 * 60000 && !order.timeout_warned) {
          orders[i].timeout_warned = true;
          orders[i].timeline.push({ status: 'timeout_warn', label: '已送达未取件超时提醒', time: new Date().toISOString(), operator: 'system' });
          modified = true;
          logger.warn('订单已送达未取件超时', { orderId: order.id, orderNo: order.order_no });
        }
      }
    }

    if (modified) {
      await writeMedicalOrdersDB(orders);
    }
  } catch (err) {
    logger.error('超时检查任务出错', { error: err.message });
  }
}

// 启动超时检查（每5分钟）
function startTimeoutChecker() {
  setInterval(checkOrderTimeouts, 5 * 60 * 1000);
  logger.info('医疗配送超时检查定时任务已启动 (每5分钟)');
}

module.exports = { router, startTimeoutChecker };
