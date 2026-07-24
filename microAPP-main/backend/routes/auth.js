/**
 * 认证路由
 */
const express = require('express');
const bcrypt = require('bcryptjs');
const jwt = require('jsonwebtoken');
const { randomUUID } = require('crypto');
const { config } = require('../config');
const { logger } = require('../logger');
const { readUsersDB, writeUsersDB } = require('../storage');
const { authRequired, rateLimit } = require('../middleware/auth');
const { sanitizeBody, requireFields } = require('../middleware/validation');
const { asyncHandler } = require('../middleware/error');
const { queryMemberByAuthCode } = require('../platformAuth');

const router = express.Router();

// 登录速率限制
router.use('/login', rateLimit({ maxRequests: 20 }));
router.use('/register', rateLimit({ maxRequests: 10 }));

/**
 * 验证用户密码
 */
async function verifyUserPassword(user, password) {
  if (!user) return false;

  // 如果有密码哈希,使用 bcrypt 验证
  if (user.passwordHash) {
    return bcrypt.compare(password, user.passwordHash);
  }

  // 向后兼容: 明文密码验证
  if (user.password) {
    const isMatch = user.password === password;
    if (isMatch) {
      // 自动升级到哈希密码
      user.passwordHash = await bcrypt.hash(password, 10);
    }
    return isMatch;
  }

  return false;
}

/**
 * 生成 Token
 */
function generateTokens(user) {
  const payload = {
    sub: user.id,
    role: user.role
  };

  const accessToken = jwt.sign(payload, config.jwt.secret, {
    expiresIn: config.jwt.accessTokenTTL
  });

  const refreshToken = jwt.sign(
    { sub: user.id, type: 'refresh' },
    config.jwt.secret,
    { expiresIn: config.jwt.refreshTokenTTL }
  );

  const decoded = jwt.decode(refreshToken);

  return {
    accessToken,
    refreshToken,
    accessTokenExpiresIn: 30 * 60, // 秒
    refreshTokenExpiresAt: decoded?.exp || 0
  };
}

/**
 * 清理用户敏感信息
 */
function sanitizeUser(user) {
  if (!user) return null;

  const {
    password,
    passwordHash,
    refreshToken,
    refreshTokenExpiresAt,
    wxSessionKey,
    ...safe
  } = user;

  return safe;
}

/**
 * 用户登录
 */
router.post('/login', sanitizeBody, requireFields('password'), asyncHandler(async (req, res) => {
  const { phone, username, password } = req.body;
  const loginId = (phone || username || '').trim();

  if (!loginId) {
    return res.status(400).json({
      success: false,
      message: '账号或密码不能为空'
    });
  }

  const users = await readUsersDB();
  const user = users.find(u => u.phone === loginId || u.username === loginId);

  if (!user || !(await verifyUserPassword(user, password))) {
    logger.warn('Login failed', { loginId, ip: req.ip });
    return res.status(401).json({
      success: false,
      message: '账号或密码错误'
    });
  }

  // 升级旧密码
  if (!user.passwordHash) {
    user.passwordHash = await bcrypt.hash(password, 10);
    await writeUsersDB(users);
  }

  const tokens = generateTokens(user);

  // 保存刷新令牌
  const updatedUsers = users.map(u => {
    if (u.id === user.id) {
      return {
        ...u,
        refreshToken: tokens.refreshToken,
        refreshTokenExpiresAt: tokens.refreshTokenExpiresAt
      };
    }
    return u;
  });

  await writeUsersDB(updatedUsers);

  logger.info('User logged in', { userId: user.id, role: user.role });

  res.json({
    success: true,
    user: sanitizeUser(user),
    ...tokens
  });
}));

/**
 * 用户注册
 */
router.post('/register', sanitizeBody, requireFields('phone', 'password', 'name'), asyncHandler(async (req, res) => {
  const { phone, password, name, username } = req.body;

  // 验证手机号格式
  const phoneRegex = /^1[3-9]\d{9}$/;
  if (!phoneRegex.test(phone)) {
    return res.status(400).json({
      success: false,
      message: '手机号格式不正确'
    });
  }

  // 验证密码强度
  if (password.length < 6) {
    return res.status(400).json({
      success: false,
      message: '密码长度不能少于6位'
    });
  }

  const users = await readUsersDB();

  // 检查手机号是否已注册
  if (users.find(u => u.phone === phone)) {
    return res.status(400).json({
      success: false,
      message: '该手机号已注册'
    });
  }

  // 创建新用户
  const newUser = {
    id: randomUUID(),
    phone,
    username: username || phone,
    password: '',
    passwordHash: await bcrypt.hash(password, 10),
    name,
    role: 'user',
    avatar: '',
    createTime: new Date().toISOString()
  };

  users.push(newUser);
  await writeUsersDB(users);

  logger.info('User registered', { userId: newUser.id, phone });

  res.json({
    success: true,
    message: '注册成功',
    user: sanitizeUser(newUser)
  });
}));

/**
 * 获取当前用户信息
 */
router.get('/me', authRequired, asyncHandler(async (req, res) => {
  const users = await readUsersDB();
  const user = users.find(u => u.id === req.user.id);

  if (!user) {
    return res.status(404).json({
      success: false,
      message: '用户不存在'
    });
  }

  res.json({
    success: true,
    user: sanitizeUser(user)
  });
}));

/**
 * 刷新 Token
 */
router.post('/refresh', asyncHandler(async (req, res) => {
  const { refreshToken } = req.body;

  if (!refreshToken) {
    return res.status(400).json({
      success: false,
      message: '缺少刷新令牌'
    });
  }

  try {
    const decoded = jwt.verify(refreshToken, config.jwt.secret);

    if (decoded.type !== 'refresh') {
      return res.status(401).json({
        success: false,
        message: '无效的令牌类型'
      });
    }

    const users = await readUsersDB();
    const user = users.find(u => u.id === decoded.sub);

    if (!user || user.refreshToken !== refreshToken) {
      return res.status(401).json({
        success: false,
        message: '刷新令牌已失效'
      });
    }

    if (Date.now() > (user.refreshTokenExpiresAt || 0) * 1000) {
      return res.status(401).json({
        success: false,
        message: '刷新令牌已过期'
      });
    }

    const tokens = generateTokens(user);

    // 更新刷新令牌
    const updatedUsers = users.map(u => {
      if (u.id === user.id) {
        return {
          ...u,
          refreshToken: tokens.refreshToken,
          refreshTokenExpiresAt: tokens.refreshTokenExpiresAt
        };
      }
      return u;
    });

    await writeUsersDB(updatedUsers);

    res.json({
      success: true,
      ...tokens
    });
  } catch (err) {
    logger.error('Token refresh failed', { error: err.message });
    res.status(401).json({
      success: false,
      message: '刷新令牌无效'
    });
  }
}));

/**
 * 登出
 */
router.post('/logout', authRequired, asyncHandler(async (req, res) => {
  const users = await readUsersDB();
  const updatedUsers = users.map(u => {
    if (u.id === req.user.id) {
      const { refreshToken, refreshTokenExpiresAt, ...rest } = u;
      return rest;
    }
    return u;
  });

  await writeUsersDB(updatedUsers);

  logger.info('User logged out', { userId: req.user.id });

  res.json({
    success: true,
    message: '登出成功'
  });
}));

/**
 * SSO 登录
 */
router.post('/sso/verify', asyncHandler(async (req, res) => {
  const { authcode } = req.body;

  if (!authcode) {
    return res.status(400).json({
      success: false,
      message: '缺少 authcode'
    });
  }

  try {
    const memberInfo = await queryMemberByAuthCode(authcode);

    if (!memberInfo || !memberInfo.phone) {
      return res.status(401).json({
        success: false,
        message: 'SSO 验证失败'
      });
    }

    const users = await readUsersDB();
    let user = users.find(u => u.phone === memberInfo.phone);

    if (!user) {
      // 自动创建用户
      user = {
        id: randomUUID(),
        phone: memberInfo.phone,
        username: memberInfo.phone,
        password: '',
        passwordHash: '',
        name: memberInfo.name || memberInfo.phone,
        role: 'user',
        avatar: memberInfo.avatar || '',
        createTime: new Date().toISOString()
      };

      users.push(user);
      await writeUsersDB(users);

      logger.info('SSO auto registered user', { userId: user.id, phone: user.phone });
    }

    const tokens = generateTokens(user);

    const updatedUsers = users.map(u => {
      if (u.id === user.id) {
        return {
          ...u,
          refreshToken: tokens.refreshToken,
          refreshTokenExpiresAt: tokens.refreshTokenExpiresAt
        };
      }
      return u;
    });

    await writeUsersDB(updatedUsers);

    res.json({
      success: true,
      user: sanitizeUser(user),
      ...tokens
    });
  } catch (err) {
    logger.error('SSO verification failed', { error: err.message });
    res.status(500).json({
      success: false,
      message: 'SSO 验证失败'
    });
  }
}));

/**
 * 获取微信公众号授权URL
 * 用于H5网页微信授权登录
 */
router.get('/wechat-oauth-url', asyncHandler(async (req, res) => {
  const { redirectUrl } = req.query;

  if (!config.wechat.mpAppId) {
    return res.status(500).json({
      success: false,
      message: '微信公众号未配置'
    });
  }

  // 微信公众号授权URL
  // scope=snsapi_userinfo: 获取用户详细信息(需要用户手动授权)
  // scope=snsapi_base: 静默授权,只获取openid
  const appId = config.wechat.mpAppId;
  const redirectUri = encodeURIComponent(`${config.server.baseUrl}/api/auth/wechat-oauth/callback`);
  const state = redirectUrl || 'home';

  const authUrl = `https://open.weixin.qq.com/connect/oauth2/authorize?appid=${appId}&redirect_uri=${redirectUri}&response_type=code&scope=snsapi_userinfo&state=${state}#wechat_redirect`;

  res.json({
    success: true,
    authUrl
  });
}));

/**
 * 微信公众号授权回调
 * 用户授权后微信会重定向到此接口,携带code参数
 */
router.get('/wechat-oauth/callback', asyncHandler(async (req, res) => {
  const { code, state } = req.query;

  if (!code) {
    return res.redirect(`${config.server.frontendUrl}/login?error=wechat_auth_failed`);
  }

  try {
    const axios = require('axios');

    // 1. 通过code获取access_token和openid
    const tokenUrl = `https://api.weixin.qq.com/sns/oauth2/access_token?appid=${config.wechat.mpAppId}&secret=${config.wechat.mpAppSecret}&code=${code}&grant_type=authorization_code`;
    const { data: tokenData } = await axios.get(tokenUrl);

    if (tokenData.errcode) {
      logger.error('WeChat OAuth failed', { error: tokenData.errmsg, code: tokenData.errcode });
      return res.redirect(`${config.server.frontendUrl}/login?error=wechat_auth_failed`);
    }

    // 2. 通过access_token和openid获取用户信息
    const userInfoUrl = `https://api.weixin.qq.com/sns/userinfo?access_token=${tokenData.access_token}&openid=${tokenData.openid}&lang=zh_CN`;
    const { data: userInfo } = await axios.get(userInfoUrl);

    if (userInfo.errcode) {
      logger.error('WeChat user info failed', { error: userInfo.errmsg });
      return res.redirect(`${config.server.frontendUrl}/login?error=wechat_auth_failed`);
    }

    // 3. 查找或创建用户
    const users = await readUsersDB();
    let user = users.find(u => u.wxOpenid === userInfo.openid);

    if (!user) {
      user = {
        id: randomUUID(),
        phone: '',
        username: `wx_mp_${userInfo.openid.substring(0, 16)}`,
        password: '',
        passwordHash: '',
        name: userInfo.nickname || '微信用户',
        role: 'user',
        avatar: userInfo.headimgurl || '',
        wxOpenid: userInfo.openid,
        wxUnionid: userInfo.unionid || '',
        createTime: new Date().toISOString()
      };

      users.push(user);
      await writeUsersDB(users);

      logger.info('WeChat MP user registered', { userId: user.id, openid: userInfo.openid });
    }

    // 4. 生成token
    const tokens = generateTokens(user);

    const updatedUsers = users.map(u => {
      if (u.id === user.id) {
        return {
          ...u,
          refreshToken: tokens.refreshToken,
          refreshTokenExpiresAt: tokens.refreshTokenExpiresAt
        };
      }
      return u;
    });

    await writeUsersDB(updatedUsers);

    // 5. 重定向到前端页面,携带token和用户信息
    const redirectPath = state && state !== 'home' ? state : 'home';
    const userData = Buffer.from(JSON.stringify(sanitizeUser(user))).toString('base64');
    const tokenDataEncoded = Buffer.from(JSON.stringify(tokens)).toString('base64');

    res.redirect(
      `${config.server.frontendUrl}/${redirectPath}?wechat_auth=1&user=${userData}&tokens=${tokenDataEncoded}`
    );
  } catch (err) {
    logger.error('WeChat OAuth callback error', { error: err.message });
    res.redirect(`${config.server.frontendUrl}/login?error=wechat_auth_error`);
  }
}));

/**
 * 微信小程序登录
 */
router.post('/wx-login', asyncHandler(async (req, res) => {
  const { code } = req.body;

  if (!code) {
    return res.status(400).json({
      success: false,
      message: '缺少登录 code'
    });
  }

  if (!config.wechat.appId || !config.wechat.appSecret) {
    return res.status(500).json({
      success: false,
      message: '微信登录未配置'
    });
  }

  // 调用微信 API
  const axios = require('axios');
  const url = `https://api.weixin.qq.com/sns/jscode2session?appid=${config.wechat.appId}&secret=${config.wechat.appSecret}&js_code=${code}&grant_type=authorization_code`;

  const { data } = await axios.get(url);

  if (data.errcode) {
    logger.error('WeChat login failed', { error: data.errmsg, code: data.errcode });
    return res.status(401).json({
      success: false,
      message: '微信登录失败'
    });
  }

  const users = await readUsersDB();
  let user = users.find(u => u.wxOpenid === data.openid);

  if (!user) {
    user = {
      id: randomUUID(),
      phone: '',
      username: `wx_${data.openid.substring(0, 16)}`,
      password: '',
      passwordHash: '',
      name: '微信用户',
      role: 'user',
      avatar: '',
      wxOpenid: data.openid,
      wxUnionid: data.unionid || '',
      createTime: new Date().toISOString()
    };

    users.push(user);
    await writeUsersDB(users);

    logger.info('WeChat user registered', { userId: user.id, openid: data.openid });
  }

  const tokens = generateTokens(user);

  const updatedUsers = users.map(u => {
    if (u.id === user.id) {
      return {
        ...u,
        refreshToken: tokens.refreshToken,
        refreshTokenExpiresAt: tokens.refreshTokenExpiresAt
      };
    }
    return u;
  });

  await writeUsersDB(updatedUsers);

  res.json({
    success: true,
    user: sanitizeUser(user),
    ...tokens
  });
}));

module.exports = router;
