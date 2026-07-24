const axios = require('axios');
const crypto = require('crypto');
const { sm2, sm4 } = require('sm-crypto');

// 畅行温州平台对接配置（测试环境）
// 来源：docs/接入文档/接入参数.txt
// 更新日期：2026-01-20
const PLATFORM_BASE_URL = process.env.PLATFORM_BASE_URL || 'https://dev.jieyisoft.com:11296';
const JOIN_INST_ID = process.env.PLATFORM_JOININST_ID || '00000015';
const H_INST_ID = process.env.PLATFORM_INST_ID || '00000001';
const MCHNT_ID = process.env.PLATFORM_MCHNT_ID || '150000150001';
const CHNL_ID = process.env.PLATFORM_CHNL_ID || '03';
const AUTH_INST_ID = process.env.PLATFORM_AUTH_INST_ID || '00000015';
const SM2_PRIVATE_KEY = process.env.PLATFORM_SM2_PRIVATE_KEY || 'cdcd3db6845a1457895328a52e109646707c6bf372ef44db69d4390989b9a5ed';

// SM4密钥处理：
// 根据文档Java示例，SM4密钥是32位字符串，直接作为hex使用（代表16字节密钥）
const SM4_KEY_RAW = process.env.PLATFORM_SM4_KEY || '12545612345648907234561434557894';
// 32位字符串直接作为hex使用，16位字符串转hex
const SM4_KEY = SM4_KEY_RAW.length === 32 
  ? SM4_KEY_RAW 
  : Buffer.from(SM4_KEY_RAW, 'utf8').toString('hex');
console.log('[SSO Config] SM4_KEY_RAW:', SM4_KEY_RAW, 'SM4_KEY:', SM4_KEY);

const platformClient = axios.create({
  baseURL: PLATFORM_BASE_URL,
  timeout: 15000
});

const ensureConfig = () => {
  if (!PLATFORM_BASE_URL) throw new Error('平台接口地址未配置');
  if (!JOIN_INST_ID) throw new Error('平台joininstid未配置');
  if (!H_INST_ID) throw new Error('平台hdata.instid未配置');
  if (!MCHNT_ID) throw new Error('平台hdata.mchntid未配置');
  if (!CHNL_ID) throw new Error('平台hdata.chnlid未配置');
  if (!SM2_PRIVATE_KEY) throw new Error('平台SM2私钥未配置');
  if (!SM4_KEY) throw new Error('平台SM4密钥未配置');
};

const formatDateTime = (date) => {
  const pad = (n) => String(n).padStart(2, '0');
  const yyyy = date.getFullYear();
  const MM = pad(date.getMonth() + 1);
  const dd = pad(date.getDate());
  const hh = pad(date.getHours());
  const mm = pad(date.getMinutes());
  const ss = pad(date.getSeconds());
  return {
    reqdate: `${yyyy}${MM}${dd}`,
    reqtime: `${hh}${mm}${ss}`
  };
};

const buildJoinInstSsn = () => {
  const now = new Date();
  const pad = (n) => String(n).padStart(2, '0');
  const ts = `${now.getFullYear()}${pad(now.getMonth() + 1)}${pad(now.getDate())}${pad(
    now.getHours()
  )}${pad(now.getMinutes())}${pad(now.getSeconds())}`;
  const rand = Math.floor(Math.random() * 1e6)
    .toString()
    .padStart(6, '0');
  return `${ts}${rand}`.slice(0, 20);
};

const buildSignContent = (payload) => {
  const keys = Object.keys(payload).sort();
  return keys.map((key) => `${key}${payload[key]}`).join('');
};

const signPayload = (payload) => {
  const signContent = buildSignContent(payload);
  console.log('[SSO Debug] signContent:', signContent);
  
  // SHA1 摘要
  const digestHex = crypto.createHash('sha1').update(signContent).digest('hex');
  console.log('[SSO Debug] SHA1 digestHex:', digestHex);
  
  // SM2 签名 - 使用 joininstid 作为 userId
  // 关键：Java 端使用 joininstidStr.getBytes("utf-8") 获取 userId 字节，
  // sm-crypto 期望 userId 为 hex 格式，因此需要将 UTF-8 字符串转为 hex
  const userIdHex = Buffer.from(JOIN_INST_ID, 'utf8').toString('hex');
  const signatureHex = sm2.doSignature(digestHex, SM2_PRIVATE_KEY, {
    hash: false,  // 我们已经做了 SHA1，不需要再 hash
    userId: userIdHex
  });
  console.log('[SSO Debug] signatureHex:', signatureHex);
  
  return Buffer.from(signatureHex, 'hex').toString('base64');
};

const encryptPayload = (plain) => {
  // 文档要求：SM4 ECB模式加密 → hex输出 → 转base64
  const plainStr = JSON.stringify(plain);
  console.log('[SSO Debug] encryptPayload plain:', plainStr);
  
  // sm4.encrypt 默认使用 ECB 模式，输出 hex 字符串
  const hexCipher = sm4.encrypt(plainStr, SM4_KEY);
  const base64Cipher = Buffer.from(hexCipher, 'hex').toString('base64');
  console.log('[SSO Debug] encrypted base64:', base64Cipher);
  
  return base64Cipher;
};

const decryptPayload = (ciphertext) => {
  // 解密：base64 → hex → SM4解密
  const hexCipher = Buffer.from(ciphertext, 'base64').toString('hex');
  const plainText = sm4.decrypt(hexCipher, SM4_KEY);
  return JSON.parse(plainText);
};

const buildRequestBody = (data) => {
  const { reqdate, reqtime } = formatDateTime(new Date());
  const hdata = {
    instid: H_INST_ID,
    mchntid: MCHNT_ID,
    chnlid: CHNL_ID
  };

  const body = {
    joininstid: JOIN_INST_ID,
    joininstssn: buildJoinInstSsn(),
    reqdate,
    reqtime,
    hdataenc: encryptPayload(hdata),
    dataenc: encryptPayload(data)
  };

  return {
    ...body,
    sign: signPayload(body)
  };
};

const postPlatform = async (path, data) => {
  ensureConfig();
  const body = buildRequestBody(data);
  
  console.log('[SSO Debug] POST', PLATFORM_BASE_URL + path);
  console.log('[SSO Debug] Request body:', JSON.stringify(body, null, 2));
  
  const res = await platformClient.post(path, body, {
    headers: {
      'Content-Type': 'application/json'
    }
  });
  
  console.log('[SSO Debug] Response:', JSON.stringify(res.data, null, 2));
  
  if (!res?.data) {
    throw new Error('平台响应为空');
  }
  if (res.data.result && res.data.result !== '0000') {
    throw new Error(res.data.resultdesc || '平台返回失败');
  }
  if (res.data.dataenc) {
    return {
      ...res.data,
      data: decryptPayload(res.data.dataenc)
    };
  }
  return res.data;
};

const queryMemberByAuthCode = async (authcode) => {
  const response = await postPlatform('/member/authaccess/member/query/V1', {
    authcode,
    authinstid: AUTH_INST_ID,
    querytype: '01'
  });
  return response?.data || {};
};

module.exports = {
  queryMemberByAuthCode
};
