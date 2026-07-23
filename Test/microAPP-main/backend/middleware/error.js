/**
 * 错误处理中间件
 */
const { logger } = require('../logger');

/**
 * 全局错误处理
 */
function errorHandler(err, req, res, next) {
  // 记录错误
  logger.errorWithContext(err, req);

  // 不要在生产环境暴露详细错误
  const isDevelopment = process.env.NODE_ENV === 'development';

  // Multer 文件上传错误
  if (err.code === 'LIMIT_FILE_SIZE') {
    return res.status(400).json({
      success: false,
      message: '文件过大'
    });
  }

  if (err.code === 'LIMIT_UNEXPECTED_FILE') {
    return res.status(400).json({
      success: false,
      message: '不支持的文件字段'
    });
  }

  // JWT 错误
  if (err.name === 'JsonWebTokenError') {
    return res.status(401).json({
      success: false,
      message: '无效的认证令牌'
    });
  }

  if (err.name === 'TokenExpiredError') {
    return res.status(401).json({
      success: false,
      message: '令牌已过期',
      code: 'TOKEN_EXPIRED'
    });
  }

  // 数据库错误
  if (err.code && err.code.startsWith('23')) {
    // PostgreSQL 约束违反
    return res.status(400).json({
      success: false,
      message: '数据违反约束,请检查输入'
    });
  }

  // 默认错误响应
  res.status(err.status || 500).json({
    success: false,
    message: err.message || '服务器内部错误',
    ...(isDevelopment && { stack: err.stack, detail: err.detail })
  });
}

/**
 * 404 处理
 */
function notFoundHandler(req, res) {
  res.status(404).json({
    success: false,
    message: '接口不存在',
    path: req.originalUrl
  });
}

/**
 * 异步错误包装器
 * 避免 try-catch 嵌套
 */
function asyncHandler(fn) {
  return (req, res, next) => {
    Promise.resolve(fn(req, res, next)).catch(next);
  };
}

module.exports = {
  errorHandler,
  notFoundHandler,
  asyncHandler
};
