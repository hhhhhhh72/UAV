/**
 * 认证和权限中间件
 */
const jwt = require('jsonwebtoken');
const { config } = require('../config');
const { logger } = require('../logger');

/**
 * 验证 JWT Token 有效性
 */
function authRequired(req, res, next) {
  const authHeader = req.headers.authorization;

  if (!authHeader || !authHeader.startsWith('Bearer ')) {
    return res.status(401).json({
      success: false,
      message: '未提供认证令牌'
    });
  }

  const token = authHeader.split(' ')[1];

  try {
    const decoded = jwt.verify(token, config.jwt.secret);
    req.user = {
      id: decoded.sub,
      role: decoded.role
    };
    next();
  } catch (err) {
    if (err.name === 'TokenExpiredError') {
      return res.status(401).json({
        success: false,
        message: '令牌已过期',
        code: 'TOKEN_EXPIRED'
      });
    }

    logger.error('JWT verification failed', { error: err.message });
    return res.status(401).json({
      success: false,
      message: '无效的认证令牌'
    });
  }
}

/**
 * 验证用户角色
 */
function roleRequired(allowedRoles) {
  return (req, res, next) => {
    if (!req.user) {
      return res.status(401).json({
        success: false,
        message: '未认证'
      });
    }

    if (!allowedRoles.includes(req.user.role)) {
      logger.warn('Unauthorized role access attempt', {
        userId: req.user.id,
        role: req.user.role,
        requiredRoles: allowedRoles,
        path: req.originalUrl
      });

      return res.status(403).json({
        success: false,
        message: '权限不足'
      });
    }

    next();
  };
}

/**
 * 可选认证 - 如果有 token 则验证,没有则跳过
 */
function optionalAuth(req, res, next) {
  const authHeader = req.headers.authorization;

  if (!authHeader || !authHeader.startsWith('Bearer ')) {
    return next();
  }

  const token = authHeader.split(' ')[1];

  try {
    const decoded = jwt.verify(token, config.jwt.secret);
    req.user = {
      id: decoded.sub,
      role: decoded.role
    };
  } catch (err) {
    // Token 无效,但不阻塞请求
    logger.debug('Optional auth failed', { error: err.message });
  }

  next();
}

/**
 * 请求速率限制 (简单实现)
 */
const rateLimitMap = new Map();

function rateLimit(options = {}) {
  const {
    windowMs = 15 * 60 * 1000, // 15分钟
    maxRequests = 100
  } = options;

  return (req, res, next) => {
    const ip = req.ip || req.socket?.remoteAddress || 'unknown';
    const now = Date.now();
    const windowStart = now - windowMs;

    if (!rateLimitMap.has(ip)) {
      rateLimitMap.set(ip, []);
    }

    const requests = rateLimitMap.get(ip).filter(time => time > windowStart);

    if (requests.length >= maxRequests) {
      logger.warn('Rate limit exceeded', { ip, requests: requests.length });
      return res.status(429).json({
        success: false,
        message: '请求过于频繁,请稍后重试'
      });
    }

    requests.push(now);
    rateLimitMap.set(ip, requests);

    // 清理旧数据
    if (requests.length < 10) {
      rateLimitMap.set(ip, requests);
    }

    next();
  };
}

/**
 * 定期清理速率限制数据
 */
setInterval(() => {
  const now = Date.now();
  const windowMs = 15 * 60 * 1000;

  for (const [ip, requests] of rateLimitMap.entries()) {
    const recent = requests.filter(time => time > now - windowMs);
    if (recent.length === 0) {
      rateLimitMap.delete(ip);
    } else {
      rateLimitMap.set(ip, recent);
    }
  }
}, 60 * 1000); // 每分钟清理一次

module.exports = {
  authRequired,
  roleRequired,
  optionalAuth,
  rateLimit
};
