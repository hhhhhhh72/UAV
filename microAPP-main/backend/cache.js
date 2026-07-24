/**
 * 缓存模块
 * 提供简单的内存缓存,减少JSON文件读取
 */
class Cache {
  constructor(options = {}) {
    this.store = new Map();
    this.defaultTTL = options.defaultTTL || 60 * 1000; // 默认1分钟
  }

  /**
   * 获取缓存
   */
  get(key) {
    const item = this.store.get(key);

    if (!item) {
      return null;
    }

    // 检查是否过期
    if (Date.now() > item.expiresAt) {
      this.store.delete(key);
      return null;
    }

    return item.value;
  }

  /**
   * 设置缓存
   */
  set(key, value, ttl = this.defaultTTL) {
    this.store.set(key, {
      value,
      expiresAt: Date.now() + ttl
    });
  }

  /**
   * 删除缓存
   */
  delete(key) {
    this.store.delete(key);
  }

  /**
   * 清空所有缓存
   */
  clear() {
    this.store.clear();
  }

  /**
   * 获取或设置(带回调)
   */
  async getOrSet(key, fn, ttl = this.defaultTTL) {
    const cached = this.get(key);

    if (cached !== null) {
      return cached;
    }

    const value = await fn();
    this.set(key, value, ttl);
    return value;
  }

  /**
   * 清理过期缓存
   */
  cleanup() {
    const now = Date.now();
    for (const [key, item] of this.store.entries()) {
      if (now > item.expiresAt) {
        this.store.delete(key);
      }
    }
  }

  /**
   * 获取缓存统计信息
   */
  stats() {
    let expired = 0;
    let active = 0;
    const now = Date.now();

    for (const item of this.store.values()) {
      if (now > item.expiresAt) {
        expired++;
      } else {
        active++;
      }
    }

    return { total: this.store.size, active, expired };
  }
}

// 创建全局缓存实例
const cache = new Cache({ defaultTTL: 60 * 1000 });

// 每5分钟清理一次过期缓存
setInterval(() => cache.cleanup(), 5 * 60 * 1000);

// 缓存键常量
const CacheKeys = {
  USERS: 'db:users',
  APPLICATIONS: 'db:applications',
  CASES: 'db:cases',
  CASE_CATEGORIES: 'db:case_categories',
  SERVICES_CONFIG: 'db:services_config',
  REVIEWS: 'db:reviews',
  MEDICAL_ORDERS: 'db:medical_orders',
  MEDICAL_CERTIFICATIONS: 'db:medical_certifications',
  MEDICAL_PADS: 'db:medical_pads',
  MEDICAL_CONTACTS: 'db:medical_contacts',
  MEDICAL_RATINGS: 'db:medical_ratings',
  MEDICAL_SMS_LOGS: 'db:medical_sms_logs',
  ADMIN_STATS: 'stats:admin',
  USER_INFO: (userId) => `user:${userId}`
};

module.exports = { Cache, cache, CacheKeys };
