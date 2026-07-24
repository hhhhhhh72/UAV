/**
 * 安全配置管理
 * 集中管理所有敏感配置,避免硬编码
 */
const crypto = require('crypto');

// 从环境变量读取配置,提供安全的默认值
const config = {
  // JWT 配置
  jwt: {
    secret: process.env.JWT_SECRET || 'change-this-to-random-string-in-production',
    accessTokenTTL: process.env.ACCESS_TOKEN_TTL || '30m',
    refreshTokenTTL: process.env.REFRESH_TOKEN_TTL || '7d'
  },

  // 管理员配置
  admin: {
    superAdminPhone: process.env.SUPER_ADMIN_PHONE || 'wzdkjjfzyxgs',
    // 管理员默认密码(首次使用后必须修改)
    defaultPasswords: {
      dslAdmin: process.env.DSL_ADMIN_PASSWORD || crypto.randomBytes(16).toString('hex'),
      studyAdmin: process.env.STUDY_ADMIN_PASSWORD || crypto.randomBytes(16).toString('hex')
    }
  },

  // 微信配置
  wechat: {
    // 小程序配置
    appId: process.env.WX_APPID || '',
    appSecret: process.env.WX_SECRET || '',
    // 公众号配置(用于H5网页授权登录)
    mpAppId: process.env.WX_MP_APPID || '',
    mpAppSecret: process.env.WX_MP_SECRET || ''
  },

  // SSO 配置
  sso: {
    apiUrl: process.env.SSO_API_URL || '',
    apiKey: process.env.SSO_API_KEY || ''
  },

  // 数据库配置
  database: {
    usePostgres: process.env.USE_POSTGRES === '1',
    pg: {
      host: process.env.PG_HOST || '127.0.0.1',
      port: parseInt(process.env.PG_PORT) || 5432,
      user: process.env.PG_USER || 'postgres',
      password: process.env.PG_PASSWORD || '',
      database: process.env.PG_DATABASE || 'lowaltitude',
      ssl: process.env.PG_SSL === 'true'
    }
  },

  // 文件上传配置
  upload: {
    dir: process.env.UPLOAD_DIR || './uploads',
    maxSize: parseInt(process.env.MAX_FILE_SIZE) || 50 * 1024 * 1024, // 50MB
    allowedTypes: (process.env.ALLOWED_FILE_TYPES || 'image/jpeg,image/png,image/gif,video/mp4').split(',')
  },

  // 服务器配置
  server: {
    port: parseInt(process.env.PORT) || 3000,
    env: process.env.NODE_ENV || 'development',
    corsOrigin: process.env.CORS_ORIGIN || 'http://localhost:5173',
    baseUrl: process.env.BASE_URL || 'http://localhost:3000',
    frontendUrl: process.env.FRONTEND_URL || 'http://localhost:5173'
  },

  // SM2 加密配置
  sm2: {
    privateKey: process.env.SM2_PRIVATE_KEY || '',
    publicKey: process.env.SM2_PUBLIC_KEY || ''
  }
};

// 验证关键配置
function validateConfig() {
  const warnings = [];

  if (config.jwt.secret === 'change-this-to-random-string-in-production') {
    warnings.push('WARNING: JWT_SECRET is using default value. Set a strong secret in production!');
  }

  if (!config.wechat.appId || !config.wechat.appSecret) {
    warnings.push('WARNING: WeChat app credentials not configured. WeChat login will not work.');
  }

  if (!config.wechat.mpAppId || !config.wechat.mpAppSecret) {
    warnings.push('WARNING: WeChat MP credentials not configured. WeChat MP OAuth will not work.');
  }

  if (config.server.env === 'production') {
    if (config.jwt.secret.length < 32) {
      throw new Error('JWT_SECRET must be at least 32 characters in production');
    }
  }

  return {
    valid: warnings.length === 0 || config.server.env !== 'production',
    warnings
  };
}

// 打印配置信息(隐藏敏感信息)
function printConfig() {
  console.log('=== Configuration ===');
  console.log(`Environment: ${config.server.env}`);
  console.log(`Port: ${config.server.port}`);
  console.log(`Database: ${config.database.usePostgres ? 'PostgreSQL' : 'JSON File'}`);
  console.log(`JWT Secret: ${config.jwt.secret.substring(0, 8)}...`);
  console.log(`WeChat AppID: ${config.wechat.appId ? 'configured' : 'not set'}`);
  console.log(`WeChat MP AppID: ${config.wechat.mpAppId ? 'configured' : 'not set'}`);
  console.log('=====================');
}

module.exports = {
  config,
  validateConfig,
  printConfig
};
