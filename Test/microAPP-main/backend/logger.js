/**
 * 日志模块
 * 使用 winston 提供结构化日志记录
 */
const fs = require('fs');
const path = require('path');

// 简单的日志实现,不依赖外部库
class Logger {
  constructor(options = {}) {
    this.logDir = options.logDir || './logs';
    this.level = options.level || 'info';
    this.ensureLogDir();
  }

  ensureLogDir() {
    if (!fs.existsSync(this.logDir)) {
      fs.mkdirSync(this.logDir, { recursive: true });
    }
  }

  getTimestamp() {
    return new Date().toISOString();
  }

  format(level, message, meta = {}) {
    const timestamp = this.getTimestamp();
    const metaStr = Object.keys(meta).length > 0 ? ` ${JSON.stringify(meta)}` : '';
    return `[${timestamp}] [${level.toUpperCase()}] ${message}${metaStr}`;
  }

  write(level, message, meta = {}) {
    const logLevels = { error: 0, warn: 1, info: 2, debug: 3 };
    const currentLevel = logLevels[this.level] || 2;

    if (logLevels[level] <= currentLevel) {
      const formatted = this.format(level, message, meta);

      // 控制台输出
      if (level === 'error') {
        console.error(formatted);
      } else if (level === 'warn') {
        console.warn(formatted);
      } else {
        console.log(formatted);
      }

      // 文件输出
      try {
        const dateStr = new Date().toISOString().split('T')[0];
        const logFile = path.join(this.logDir, `${dateStr}.log`);
        fs.appendFileSync(logFile, formatted + '\n');
      } catch (err) {
        console.error('Failed to write log file:', err.message);
      }
    }
  }

  info(message, meta = {}) {
    this.write('info', message, meta);
  }

  warn(message, meta = {}) {
    this.write('warn', message, meta);
  }

  error(message, meta = {}) {
    this.write('error', message, meta);
  }

  debug(message, meta = {}) {
    this.write('debug', message, meta);
  }

  // 请求日志
  request(req, responseTime) {
    this.info(`${req.method} ${req.originalUrl || req.url}`, {
      ip: req.ip || req.socket?.remoteAddress,
      status: req.statusCode,
      responseTime: `${responseTime}ms`,
      userAgent: req.headers['user-agent']
    });
  }

  // 错误日志
  errorWithContext(err, req) {
    this.error(`${err.message || 'Unknown error'}`, {
      method: req.method,
      url: req.originalUrl || req.url,
      ip: req.ip || req.socket?.remoteAddress,
      stack: err.stack,
      body: req.body ? JSON.stringify(req.body).substring(0, 500) : undefined
    });
  }
}

// 创建默认日志实例
const logger = new Logger({
  logDir: process.env.LOG_DIR || './logs',
  level: process.env.LOG_LEVEL || 'info'
});

module.exports = { Logger, logger };
