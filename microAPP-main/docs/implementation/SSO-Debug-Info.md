# 畅行温州SSO对接调试信息

## 当前配置参数

| 参数 | 值 |
|------|-----|
| 请求地址 | https://dev.jieyisoft.com:11296 |
| joininstid | 00000015 |
| hdata.instid | 00000001 |
| hdata.mchntid | 150000150001 |
| hdata.chnlid | 03 |
| authinstid | 00000015 |
| SM2 私钥 | cdcd3db6845a1457895328a52e109646707c6bf372ef44db69d4390989b9a5ed |
| SM2 公钥 | 04b0820c10227017749f6ea25ef49f0d65cf53f701a0230179362470eaf45a560c3caf88558d03c0e74fad2c78cfeb1aa46e1c232d7519371ae0e33057fd0e66c3 |
| SM4 密钥 | 12545612345648907234561434557894 (32位hex) |

## 签名流程验证

### 1. 签名内容构建

按ASCII顺序排列key并拼接：
```
dataenc{加密值}hdataenc{加密值}joininstid{值}joininstssn{值}reqdate{值}reqtime{值}
```

**示例：**
```
dataencwlEXxJXGiMPI5Sw73to+6F9rYnvMU80vWKl9jWDq0g1r4RvqHOkd/WI7eykSrcPm9J4vdrHUuO1wUCLX01e9/Q==hdataencLeounLGOdgl568M2ChqK3o+JaLHewez8aqsUjqMtoYRs0WCuBEx2RXLsgnVfLhvqB5b47fgU59WP4sZ8wePlIA==joininstid00000015joininstssn20260121151547437102reqdate20260121reqtime151547
```

### 2. SHA1摘要

对签名内容进行SHA1哈希：
```
SHA1摘要: 12db3ed8a6cc0ea1a0786771680d071aedaaed2b
```

### 3. SM2签名

使用私钥对SHA1摘要进行SM2签名：
- `hash: false` - 已经做过SHA1，不需要再hash
- `userId: joininstid` - 即 "00000015"
- 输出为Base64格式

## 实际请求示例

```json
{
  "joininstid": "00000015",
  "joininstssn": "20260121151547437102",
  "reqdate": "20260121",
  "reqtime": "151547",
  "hdataenc": "LeounLGOdgl568M2ChqK3o+JaLHewez8aqsUjqMtoYRs0WCuBEx2RXLsgnVfLhvqB5b47fgU59WP4sZ8wePlIA==",
  "dataenc": "wlEXxJXGiMPI5Sw73to+6F9rYnvMU80vWKl9jWDq0g1r4RvqHOkd/WI7eykSrcPm9J4vdrHUuO1wUCLX01e9/Q==",
  "sign": "/AQJ4bierstfhHHb4ryn9t+68cBKWr8MZhyq34w6atCGz1DQUMEGiejPiKTrHFZnDldYrIf8Ue1OBbQvJnJhwA=="
}
```

## 平台响应

```json
{
  "result": "SB997",
  "resultdesc": "签名错误"
}
```

## 本地验证结果

1. **密钥对验证** ✅ 
   - 使用私钥签名后，可以用公钥成功验签
   - 说明密钥对是配对的

2. **SHA1计算验证** ✅
   - 与文档示例计算结果一致

3. **SM4加密验证** ✅
   - 使用16字符ASCII密钥转hex后加密，结果与文档示例一致

## 需要确认的问题

请平台开发团队帮助确认：

1. **公钥是否已正确注册？**
   - 我方公钥: `04b0820c10227017749f6ea25ef49f0d65cf53f701a0230179362470eaf45a560c3caf88558d03c0e74fad2c78cfeb1aa46e1c232d7519371ae0e33057fd0e66c3`
   - 对应joininstid: `00000015`

2. **SM2签名参数是否一致？**
   - userId: 我们使用 `joininstid` 的值 ("00000015") 作为SM2签名的userId
   - 签名输入: SHA1摘要的hex字符串
   - 签名格式: raw r||s (64字节/128字符hex)，转Base64

3. **能否提供验签失败的具体原因？**
   - 是公钥不匹配？
   - 是签名格式不对？
   - 还是签名内容构建方式不对？

## 签名代码参考 (Node.js)

```javascript
const crypto = require('crypto');
const { sm2 } = require('sm-crypto');

// 构建签名内容
const signContent = Object.keys(body)
  .sort()
  .map(key => `${key}${body[key]}`)
  .join('');

// SHA1摘要
const sha1Hex = crypto.createHash('sha1').update(signContent).digest('hex');

// SM2签名
const signature = sm2.doSignature(sha1Hex, privateKey, {
  hash: false,
  userId: joininstid
});

// 转Base64
const signBase64 = Buffer.from(signature, 'hex').toString('base64');
```

## 日期

2026-01-21
