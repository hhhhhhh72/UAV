/**
 * SM2 签名工具
 * 提供与 Java SDK 兼容的 SM2 签名和验签
 */
const { sm2 } = require('sm-crypto');
const { config } = require('../config');
const { logger } = require('../logger');

/**
 * SM2 签名
 *
 * Java SDK 调用方式:
 * SM2Utils.sign(userId.getBytes(), privateKey_bytes, sourceData_bytes)
 *
 * @param {string} message - 待签名的消息(hex字符串或UTF-8字符串)
 * @param {string} userId - 用户ID(用于Z值计算)
 * @param {Object} options - 选项
 * @param {boolean} options.hash - 是否对消息进行hash,默认false
 * @param {boolean} options.der - 是否使用DER格式,默认true
 * @returns {string} 签名结果(hex字符串)
 */
function sign(message, userId, options = {}) {
  const {
    hash = false,
    der = true
  } = options;

  try {
    const privateKey = config.sm2.privateKey;

    if (!privateKey) {
      throw new Error('SM2 私钥未配置');
    }

    let msgHex = message;

    // 如果消息不是hex格式,先转换为hex
    if (!/^[0-9a-fA-F]+$/.test(message)) {
      msgHex = Buffer.from(message, 'utf8').toString('hex');
    }

    // userId 也转换为hex
    const userIdHex = userId ? Buffer.from(userId, 'utf8').toString('hex') : '';

    // 执行签名
    const signature = sm2.doSignature(msgHex, privateKey, {
      hash,
      userId: userIdHex,
      der
    });

    logger.debug('SM2 signature generated', {
      messageLength: message.length,
      signatureLength: signature.length
    });

    return signature;
  } catch (error) {
    logger.error('SM2 sign failed', { error: error.message });
    throw error;
  }
}

/**
 * SM2 验签
 *
 * @param {string} message - 原始消息(hex字符串或UTF-8字符串)
 * @param {string} signature - 签名(hex字符串)
 * @param {string} userId - 用户ID
 * @param {Object} options - 选项
 * @param {boolean} options.hash - 是否对消息进行hash,默认false
 * @param {boolean} options.der - 是否使用DER格式,默认true
 * @returns {boolean} 验签结果
 */
function verify(message, signature, userId, options = {}) {
  const {
    hash = false,
    der = true
  } = options;

  try {
    const publicKey = config.sm2.publicKey;

    if (!publicKey) {
      throw new Error('SM2 公钥未配置');
    }

    let msgHex = message;

    // 如果消息不是hex格式,先转换为hex
    if (!/^[0-9a-fA-F]+$/.test(message)) {
      msgHex = Buffer.from(message, 'utf8').toString('hex');
    }

    // userId 也转换为hex
    const userIdHex = userId ? Buffer.from(userId, 'utf8').toString('hex') : '';

    // 执行验签
    const result = sm2.doVerifySignature(msgHex, signature, publicKey, {
      hash,
      userId: userIdHex,
      der
    });

    logger.debug('SM2 signature verification', { result });

    return result;
  } catch (error) {
    logger.error('SM2 verify failed', { error: error.message });
    return false;
  }
}

/**
 * 生成 SM2 密钥对
 *
 * @returns {Object} 包含私钥和公钥的对象
 */
function generateKeyPair() {
  const keypair = sm2.generateKeyPairHex();
  return {
    privateKey: keypair.privateKey,
    publicKey: keypair.publicKey
  };
}

/**
 * 针对 Java SDK 的兼容签名
 *
 * Java 代码:
 * SM2Utils.sign(joininstidStr.getBytes("utf-8"), Util.hexToByte(sendPrivatekey), Util.hexToByte(strForSignSha1))
 *
 * @param {string} userId - joininstid (UTF-8字符串)
 * @param {string} privateKeyHex - 私钥(hex字符串)
 * @param {string} sourceDataHex - 已hash的数据(hex字符串,如SHA1结果)
 * @returns {string} 签名(base64字符串)
 */
function signForJavaSDK(userId, privateKeyHex, sourceDataHex) {
  try {
    // userId 转换为hex
    const userIdHex = Buffer.from(userId, 'utf8').toString('hex');

    // 使用已hash的数据进行签名(hash=false)
    const signature = sm2.doSignature(sourceDataHex, privateKeyHex, {
      hash: false,
      userId: userIdHex,
      der: false  // Java SDK 默认使用r||s格式
    });

    // 转换为base64
    return Buffer.from(signature, 'hex').toString('base64');
  } catch (error) {
    logger.error('SM2 sign for Java SDK failed', { error: error.message });
    throw error;
  }
}

/**
 * 针对 Java SDK 的兼容验签
 *
 * @param {string} userId - joininstid (UTF-8字符串)
 * @param {string} publicKeyHex - 公钥(hex字符串)
 * @param {string} sourceDataHex - 已hash的数据(hex字符串)
 * @param {string} signatureBase64 - 签名(base64字符串)
 * @returns {boolean} 验签结果
 */
function verifyForJavaSDK(userId, publicKeyHex, sourceDataHex, signatureBase64) {
  try {
    const userIdHex = Buffer.from(userId, 'utf8').toString('hex');
    const signatureHex = Buffer.from(signatureBase64, 'base64').toString('hex');

    return sm2.doVerifySignature(sourceDataHex, signatureHex, publicKeyHex, {
      hash: false,
      userId: userIdHex,
      der: false
    });
  } catch (error) {
    logger.error('SM2 verify for Java SDK failed', { error: error.message });
    return false;
  }
}

module.exports = {
  sign,
  verify,
  generateKeyPair,
  signForJavaSDK,
  verifyForJavaSDK
};
