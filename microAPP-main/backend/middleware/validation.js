/**
 * 输入验证和消毒中间件
 */
const { logger } = require('../logger');

/**
 * 消毒字符串输入
 */
function sanitizeString(str) {
  if (typeof str !== 'string') return str;

  // 移除潜在的 XSS 脚本标签
  let sanitized = str.replace(/<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi, '');

  // 移除 HTML 事件属性
  sanitized = sanitized.replace(/\s+on\w+\s*=\s*["'][^"]*["']/gi, '');

  // 移除 javascript: 协议
  sanitized = sanitized.replace(/javascript\s*:/gi, '');

  return sanitized.trim();
}

/**
 * 消毒对象的所有字符串字段
 */
function sanitizeObject(obj) {
  if (!obj || typeof obj !== 'object') return obj;

  const sanitized = {};

  for (const [key, value] of Object.entries(obj)) {
    if (typeof value === 'string') {
      sanitized[key] = sanitizeString(value);
    } else if (Array.isArray(value)) {
      sanitized[key] = value.map(item =>
        typeof item === 'string' ? sanitizeString(item) : item
      );
    } else if (typeof value === 'object' && value !== null) {
      sanitized[key] = sanitizeObject(value);
    } else {
      sanitized[key] = value;
    }
  }

  return sanitized;
}

/**
 * 请求体消毒中间件
 */
function sanitizeBody(req, res, next) {
  if (req.body) {
    req.body = sanitizeObject(req.body);
  }
  next();
}

/**
 * 验证手机号格式
 */
function validatePhone(phone) {
  // 中国手机号格式: 11位数字,以1开头
  const phoneRegex = /^1[3-9]\d{9}$/;
  return phoneRegex.test(phone);
}

/**
 * 验证邮箱格式
 */
function validateEmail(email) {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
}

/**
 * 验证必填字段
 */
function requireFields(...fields) {
  return (req, res, next) => {
    const missing = fields.filter(field => {
      const value = field.includes('.')
        ? field.split('.').reduce((obj, key) => obj?.[key], req.body)
        : req.body[field];

      return !value || (typeof value === 'string' && !value.trim());
    });

    if (missing.length > 0) {
      return res.status(400).json({
        success: false,
        message: `缺少必填字段: ${missing.join(', ')}`
      });
    }

    next();
  };
}

/**
 * 验证并清理查询参数
 */
function validateQuery(schema) {
  return (req, res, next) => {
    const errors = [];

    for (const [field, rules] of Object.entries(schema)) {
      const value = req.query[field];

      if (rules.required && !value) {
        errors.push(`缺少参数: ${field}`);
        continue;
      }

      if (value) {
        if (rules.type === 'number') {
          const num = Number(value);
          if (isNaN(num)) {
            errors.push(`参数 ${field} 必须是数字`);
          } else if (rules.min !== undefined && num < rules.min) {
            errors.push(`参数 ${field} 不能小于 ${rules.min}`);
          } else if (rules.max !== undefined && num > rules.max) {
            errors.push(`参数 ${field} 不能大于 ${rules.max}`);
          } else {
            req.query[field] = num;
          }
        }

        if (rules.type === 'string' && rules.maxLength && value.length > rules.maxLength) {
          errors.push(`参数 ${field} 不能超过 ${rules.maxLength} 个字符`);
        }

        if (rules.enum && !rules.enum.includes(value)) {
          errors.push(`参数 ${field} 必须是以下值之一: ${rules.enum.join(', ')}`);
        }
      }
    }

    if (errors.length > 0) {
      return res.status(400).json({
        success: false,
        message: '参数验证失败',
        errors
      });
    }

    next();
  };
}

/**
 * 文件上传验证
 */
function validateFileUpload(req, res, next) {
  if (!req.file) {
    return next();
  }

  const { config } = require('../config');
  const allowedTypes = config.upload.allowedTypes;
  const maxSize = config.upload.maxSize;

  // 验证文件类型
  if (!allowedTypes.includes(req.file.mimetype)) {
    return res.status(400).json({
      success: false,
      message: `不支持的文件类型。允许的类型: ${allowedTypes.join(', ')}`
    });
  }

  // 验证文件大小
  if (req.file.size > maxSize) {
    return res.status(400).json({
      success: false,
      message: `文件过大。最大允许: ${Math.round(maxSize / 1024 / 1024)}MB`
    });
  }

  next();
}

module.exports = {
  sanitizeBody,
  sanitizeString,
  sanitizeObject,
  validatePhone,
  validateEmail,
  requireFields,
  validateQuery,
  validateFileUpload
};
