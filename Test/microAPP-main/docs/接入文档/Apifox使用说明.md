# 畅行温州 会员查询 - Apifox 使用说明

## 1. 用接入参数生成请求

接入参数见 `接入参数.txt`，已用于 `create-request.js`。

在项目根目录执行：

```bash
cd backend
node create-request.js [authcode]
```

- 不传 `authcode`：使用 `test_<时间戳>`
- 示例：`node create-request.js test123`

脚本会输出：

- **请求 URL**：`https://dev.jieyisoft.com:11296/member/authaccess/member/query/V1`
- **Headers**：`Content-Type`、`Accept`
- **Request Body**：完整 JSON（含 `joininstid`、`joininstssn`、`reqdate`、`reqtime`、`hdataenc`、`dataenc`、`sign`）
- **调试信息**：`signContent`、`sha1Hex`、`signatureHex`、`signatureBase64`

每次运行会生成新的 `joininstssn`/`reqdate`/`reqtime`，并重新计算 `hdataenc`、`dataenc`、`sign`，所以每次都要用**本次输出**的 Body。

---

## 2. 在 Apifox 里发请求

### 方式 A：导入 Postman 集合

1. Apifox → **导入** → **Postman**
2. 选择 `畅行温州-会员查询-Apifox.json`
3. 打开「会员查询 member/query/V1」请求
4. **Body** 已预填一份示例；要发新请求时，用 `create-request.js` 的输出**整体替换** Body 的 JSON

### 方式 B：手动建接口

| 项目 | 值 |
|------|-----|
| 方法 | `POST` |
| URL | `https://dev.jieyisoft.com:11296/member/authaccess/member/query/V1` |
| Headers | `Content-Type: application/json;charset=utf-8`<br>`Accept: application/json` |
| Body | `raw` → `JSON`，粘贴 `create-request.js` 输出的 **Request Body** |

---

## 3. 接入参数速查（与 create-request.js 一致）

| 参数 | 值 |
|------|-----|
| 请求地址 | `https://dev.jieyisoft.com:11296` |
| joininstid | `00000015` |
| authinstid | `00000015` |
| hdata.instid | `00000001` |
| hdata.mchntid | `150000150001` |
| hdata.chnlid | `03` |
| SM2 公钥 / 私钥 / SM4 密钥 | 见 `接入参数.txt`（仅脚本内部使用） |

---

## 4. 流程小结

1. 运行 `node create-request.js [authcode]`
2. 复制输出的 **Request Body**（或「单行 Body」）
3. 在 Apifox 的会员查询接口中，Body 选 raw JSON，粘贴并发送

这样即可用当前接入参数在 Apifox 里创建并发送畅行温州会员查询请求。
