// Package wechatpay 微信支付 APIv3（JSAPI / 小程序支付）。
//
// 设计取舍：**回调只做触发，以主动查询为准**。
//
// 微信文档要求在回调里用平台证书验签。本实现不走那条路，理由：
//  1. 平台证书会轮换，得实现下载/缓存/轮换，是一处长期运维负担；
//  2. 收到回调只用 APIv3 密钥解密拿到商户订单号，**随即主动回查微信**
//     「这笔到底付没付、付了多少」，以回查结果入账。回查由我们发起、走 TLS
//     直连 api.mch.weixin.qq.com，结果可信；
//  3. 于是伪造回调最多触发一次返回「未支付」的查询，不产生任何资金影响。
//
// 关键红线：**金额只信回查结果**，并与本地订单金额逐分比对。
package wechatpay

import (
	"bytes"
	"context"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// effectiveRefundNotifyURL 返回退款回调地址：显式配置优先，否则由 NotifyURL 推导
// （/wechat/notify → /wechat/refund-notify）。
func (c Config) effectiveRefundNotifyURL() string {
	if c.RefundNotifyURL != "" {
		return c.RefundNotifyURL
	}
	if c.NotifyURL == "" {
		return ""
	}
	if strings.HasSuffix(c.NotifyURL, "/notify") {
		return strings.TrimSuffix(c.NotifyURL, "/notify") + "/refund-notify"
	}
	return c.NotifyURL
}

// defaultAPIBase 微信支付 APIv3 正式域名。注意这里**不能**带反引号：
// 之前写成 "`https://api.mch.weixin.qq.com`"，字符串里真的多出两个反引号，
// 而 Config.APIBase 只有测试会覆盖 → 生产每一笔下单一律 DNS 解析失败。
// 由 TestDefaultAPIBaseIsCleanURL 守住。
const defaultAPIBase = "https://api.mch.weixin.qq.com"

// Config 客户端配置。五项齐全才可 New（与 config.WeChatPayConfig 对齐）。
type Config struct {
	AppID          string
	MchID          string
	APIv3Key       string // 32 字节，回调解密用
	CertSerial     string // 商户 API 证书序列号
	PrivateKeyPath string // apiclient_key.pem
	// NotifyURL 支付结果回调地址。
	NotifyURL string
	// RefundNotifyURL 退款结果回调地址。留空时由 NotifyURL 推导（.../notify →
	// .../refund-notify），便于只配一个域名就把两条链路都接上。
	RefundNotifyURL string
	// APIBase 仅测试用：指向本地 httptest 服务器。留空走线上域名。
	APIBase string
	// HTTPClient 留空用带 10s 超时的默认客户端。
	HTTPClient *http.Client
}

type Client struct {
	cfg     Config
	priv    *rsa.PrivateKey
	http    *http.Client
	baseURL string
}

// LoadPrivateKeyFromPEM 读商户私钥，PKCS#1（RSA PRIVATE KEY）与 PKCS#8（PRIVATE KEY）都认。
// 微信商户平台下载的 apiclient_key.pem 是 PKCS#8，早期工具产出的是 PKCS#1。
func LoadPrivateKeyFromPEM(pemBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("wechatpay: 私钥不是合法 PEM")
	}
	if k, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return k, nil
	}
	anyKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("wechatpay: 解析私钥失败: %w", err)
	}
	k, ok := anyKey.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("wechatpay: 私钥不是 RSA 类型（微信支付要求 RSA）")
	}
	return k, nil
}

// New 构造客户端。缺项或密钥不可用时立即失败——启动即暴露，不留到收钱那一刻。
func New(cfg Config) (*Client, error) {
	if cfg.MchID == "" || cfg.APIv3Key == "" || cfg.CertSerial == "" || cfg.PrivateKeyPath == "" || cfg.NotifyURL == "" {
		return nil, errors.New("wechatpay: 配置不完整（商户号/APIv3密钥/证书序列号/私钥路径/回调地址五项必填）")
	}
	if len(cfg.APIv3Key) != 32 {
		return nil, fmt.Errorf("wechatpay: APIv3 密钥必须 32 字节，当前 %d", len(cfg.APIv3Key))
	}
	raw, err := os.ReadFile(cfg.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("wechatpay: 读取商户私钥失败: %w", err)
	}
	priv, err := LoadPrivateKeyFromPEM(raw)
	if err != nil {
		return nil, err
	}
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	base := cfg.APIBase
	if base == "" {
		base = defaultAPIBase
	}
	return &Client{cfg: cfg, priv: priv, http: httpClient, baseURL: strings.TrimRight(base, "/")}, nil
}

// nonce 生成 32 位十六进制随机串（微信要求不长于 32 位）。
func nonce() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("wechatpay: 生成随机串失败: %w", err)
	}
	return fmt.Sprintf("%x", b), nil
}

// signature 用商户私钥签名（SHA256withRSA），返回 base64。
func (c *Client) signature(message string) (string, error) {
	digest := sha256.Sum256([]byte(message))
	sig, err := rsa.SignPKCS1v15(rand.Reader, c.priv, crypto.SHA256, digest[:])
	if err != nil {
		return "", fmt.Errorf("wechatpay: 签名失败: %w", err)
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

// authMessage 微信 v3 请求签名串：METHOD\nURL(path+query)\n时间戳\n随机串\n请求体\n
// 每一行（含最后一行）都以换行结尾——漏掉末尾换行会得到 401 Signature verify fail。
func authMessage(method, urlPath, body, timestamp, nonceStr string) string {
	return method + "\n" + urlPath + "\n" + timestamp + "\n" + nonceStr + "\n" + body + "\n"
}

// authorization 组装 Authorization 头。
func (c *Client) authorization(method, urlPath, body string) (string, error) {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	n, err := nonce()
	if err != nil {
		return "", err
	}
	sig, err := c.signature(authMessage(method, urlPath, body, ts, n))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`WECHATPAY2-SHA256-RSA2048 mchid="%s",nonce_str="%s",signature="%s",timestamp="%s",serial_no="%s"`,
		c.cfg.MchID, n, sig, ts, c.cfg.CertSerial), nil
}

// APIError 微信返回的业务错误（HTTP 非 2xx）。
type APIError struct {
	StatusCode int
	Code       string
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("wechatpay: API 返回 %d %s: %s", e.StatusCode, e.Code, e.Message)
}

// do 发起一次 v3 请求；成功时把响应解到 out（out 为 nil 表示不关心响应体）。
func (c *Client) do(ctx context.Context, method, urlPath string, payload, out any) error {
	var body []byte
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("wechatpay: 序列化请求失败: %w", err)
		}
		body = b
	}
	auth, err := c.authorization(method, urlPath, string(body))
	if err != nil {
		return err
	}
	var rd io.Reader
	if body != nil {
		rd = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+urlPath, rd)
	if err != nil {
		return fmt.Errorf("wechatpay: 构造请求失败: %w", err)
	}
	req.Header.Set("Authorization", auth)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "drone-platform/1.0")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("wechatpay: 请求失败: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("wechatpay: 读取响应失败: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return parseAPIError(resp.StatusCode, respBody)
	}
	if out != nil {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("wechatpay: 解析响应失败: %w", err)
		}
	}
	return nil
}

// parseAPIError 从错误响应体提取 code/message；体不是 JSON 时退化为原文。
func parseAPIError(status int, body []byte) error {
	var e struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &e); err != nil || e.Code == "" {
		return &APIError{StatusCode: status, Code: "UNKNOWN", Message: strings.TrimSpace(string(body))}
	}
	return &APIError{StatusCode: status, Code: e.Code, Message: e.Message}
}

// CreateJSAPIOrder 小程序下单，返回 prepay_id（前端调起支付用，**不是到账凭证**）。
func (c *Client) CreateJSAPIOrder(ctx context.Context, outTradeNo, description string, amountFen int64, payerOpenID string) (string, error) {
	if outTradeNo == "" {
		return "", errors.New("wechatpay: 商户订单号必填")
	}
	if payerOpenID == "" {
		return "", errors.New("wechatpay: payer.openid 必填（小程序支付必须知道谁付钱）")
	}
	if amountFen <= 0 {
		return "", fmt.Errorf("wechatpay: 金额必须为正，当前 %d", amountFen)
	}
	var in struct {
		AppID       string `json:"appid"`
		MchID       string `json:"mchid"`
		Description string `json:"description"`
		OutTradeNo  string `json:"out_trade_no"`
		NotifyURL   string `json:"notify_url"`
		Amount      struct {
			Total    int64  `json:"total"`
			Currency string `json:"currency"`
		} `json:"amount"`
		Payer struct {
			OpenID string `json:"openid"`
		} `json:"payer"`
	}
	in.AppID = c.cfg.AppID
	in.MchID = c.cfg.MchID
	in.Description = description
	in.OutTradeNo = outTradeNo
	in.NotifyURL = c.cfg.NotifyURL
	in.Amount.Total = amountFen
	in.Amount.Currency = "CNY"
	in.Payer.OpenID = payerOpenID
	var out struct {
		PrepayID string `json:"prepay_id"`
	}
	if err := c.do(ctx, http.MethodPost, "/v3/pay/transactions/jsapi", in, &out); err != nil {
		return "", err
	}
	if out.PrepayID == "" {
		return "", errors.New("wechatpay: 下单成功但未返回 prepay_id")
	}
	return out.PrepayID, nil
}

// TradeStateSuccess 微信侧「支付成功」。其余状态（NOTPAY/CLOSED/REVOKED/USERPAYING/PAYERROR）
// 一律不得入账。
const TradeStateSuccess = "SUCCESS"

// Transaction 微信侧订单（查单返回）。这是唯一可信的到账依据。
type Transaction struct {
	AppID          string `json:"appid"`
	MchID          string `json:"mchid"`
	OutTradeNo     string `json:"out_trade_no"`
	TransactionID  string `json:"transaction_id"`
	TradeState     string `json:"trade_state"`
	TradeStateDesc string `json:"trade_state_desc"`
	SuccessTime    string `json:"success_time"`
	Amount         struct {
		Total    int64  `json:"total"`
		Currency string `json:"currency"`
	} `json:"amount"`
}

// QueryOrder 主动查单。
func (c *Client) QueryOrder(ctx context.Context, outTradeNo string) (*Transaction, error) {
	if outTradeNo == "" {
		return nil, errors.New("wechatpay: 商户订单号必填")
	}
	path := "/v3/pay/transactions/out-trade-no/" + url.PathEscape(outTradeNo) + "?mchid=" + url.QueryEscape(c.cfg.MchID)
	var tx Transaction
	if err := c.do(ctx, http.MethodGet, path, nil, &tx); err != nil {
		return nil, err
	}
	return &tx, nil
}

// PayParams 生成 wx.requestPayment 所需参数。签名串格式与请求签名不同：
// appId\n时间戳\n随机串\nprepay_id=xxx\n
func (c *Client) PayParams(prepayID string) (map[string]string, error) {
	if prepayID == "" {
		return nil, errors.New("wechatpay: prepay_id 为空")
	}
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	n, err := nonce()
	if err != nil {
		return nil, err
	}
	pkg := "prepay_id=" + prepayID
	paySign, err := c.signature(c.cfg.AppID + "\n" + ts + "\n" + n + "\n" + pkg + "\n")
	if err != nil {
		return nil, err
	}
	return map[string]string{
		"timeStamp": ts,
		"nonceStr":  n,
		"package":   pkg,
		"signType":  "RSA",
		"paySign":   paySign,
	}, nil
}

// NotifyEnvelope 支付结果回调的外层信封（回调本身不作为到账依据，只用来取商户订单号）。
type NotifyEnvelope struct {
	ID           string `json:"id"`
	CreateTime   string `json:"create_time"`
	EventType    string `json:"event_type"`
	ResourceType string `json:"resource_type"`
	Resource     struct {
		Algorithm      string `json:"algorithm"`
		Ciphertext     string `json:"ciphertext"`
		Nonce          string `json:"nonce"`
		AssociatedData string `json:"associated_data"`
	} `json:"resource"`
}

// DecryptResource 用 APIv3 密钥做 AES-256-GCM 解密，返回明文报文。
// 算法不符/字段缺失一律报错，不做兜底猜测。
func (c *Client) DecryptResource(env NotifyEnvelope) ([]byte, error) {
	if env.Resource.Algorithm != "" && env.Resource.Algorithm != "AEAD_AES_256_GCM" {
		return nil, fmt.Errorf("wechatpay: 不支持的加密算法 %q", env.Resource.Algorithm)
	}
	if env.Resource.Ciphertext == "" || env.Resource.Nonce == "" {
		return nil, errors.New("wechatpay: 回调缺少 ciphertext 或 nonce")
	}
	raw, err := base64.StdEncoding.DecodeString(env.Resource.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("wechatpay: 回调密文不是合法 base64: %w", err)
	}
	block, err := aes.NewCipher([]byte(c.cfg.APIv3Key))
	if err != nil {
		return nil, fmt.Errorf("wechatpay: 初始化 AES 失败: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("wechatpay: 初始化 GCM 失败: %w", err)
	}
	if len(env.Resource.Nonce) != gcm.NonceSize() {
		return nil, fmt.Errorf("wechatpay: nonce 长度应为 %d，实际 %d", gcm.NonceSize(), len(env.Resource.Nonce))
	}
	plain, err := gcm.Open(nil, []byte(env.Resource.Nonce), raw, []byte(env.Resource.AssociatedData))
	if err != nil {
		return nil, fmt.Errorf("wechatpay: 回调解密失败（APIv3 密钥不符或报文被篡改）: %w", err)
	}
	return plain, nil
}
// ── 退款（APIv3 /v3/refund/domestic/refunds）──
//
// 退款的对象是**某一笔充值单**，不是某个订单：平台的钱是以「充值进托管余额」的形式进来的
// （见 domain.PaymentOrder），用户微信付的那笔钱对应一条 payment_orders 记录。
// 微信要求带 out_trade_no、且累计退款不超过原单金额，并支持多次部分退款——
// 所以「累计不超额」必须由调用方保证（仓库层用 CAS 更新 refunded_fen）。

// RefundRequest 发起一次退款。
type RefundRequest struct {
	OutTradeNo  string // 原充值单号（payment_orders.out_trade_no）
	OutRefundNo string // 本次退款单号（我们自己生成，全局唯一）
	RefundFen   int64  // 本次退款金额
	TotalFen    int64  // 原单总金额（微信要求一并带上）
	Reason      string // 退款原因（可选）
}

// Refund 微信侧退款单。
type Refund struct {
	RefundID    string `json:"refund_id"`
	OutRefundNo string `json:"out_refund_no"`
	OutTradeNo  string `json:"out_trade_no"`
	Status      string `json:"status"`
	CreateTime  string `json:"create_time"`
	SuccessTime string `json:"success_time"`
	Amount      struct {
		Total       int64 `json:"total"`
		Refund      int64 `json:"refund"`
		PayerTotal  int64 `json:"payer_total"`
		PayerRefund int64 `json:"payer_refund"`
	} `json:"amount"`
}

// 微信退款状态。只有 RefundStatusSuccess 代表钱真的退回去了。
const (
	RefundStatusSuccess    = "SUCCESS"
	RefundStatusProcessing = "PROCESSING" // 已受理、处理中（银行侧未完成）
	RefundStatusClosed     = "CLOSED"     // 已关闭（退款失败/被撤销）
	RefundStatusAbnormal   = "ABNORMAL"   // 异常，需人工介入
)

// CreateRefund 发起退款。返回成功**只代表微信受理了**，钱是否真退到账要以
// QueryRefund（或退款回调）的 status 为准——与下单只拿 prepay_id 同理。
func (c *Client) CreateRefund(ctx context.Context, in RefundRequest) (*Refund, error) {
	if in.OutTradeNo == "" || in.OutRefundNo == "" {
		return nil, errors.New("wechatpay: 原单号与退款单号必填")
	}
	if in.RefundFen <= 0 {
		return nil, fmt.Errorf("wechatpay: 退款金额必须为正，当前 %d", in.RefundFen)
	}
	if in.TotalFen < in.RefundFen {
		return nil, fmt.Errorf("wechatpay: 退款金额 %d 不得超过原单金额 %d", in.RefundFen, in.TotalFen)
	}
	var body struct {
		OutTradeNo  string `json:"out_trade_no"`
		OutRefundNo string `json:"out_refund_no"`
		Reason      string `json:"reason,omitempty"`
		NotifyURL   string `json:"notify_url,omitempty"`
		Amount      struct {
			Refund   int64  `json:"refund"`
			Total    int64  `json:"total"`
			Currency string `json:"currency"`
		} `json:"amount"`
	}
	body.OutTradeNo = in.OutTradeNo
	body.OutRefundNo = in.OutRefundNo
	body.Reason = in.Reason
	body.NotifyURL = c.cfg.effectiveRefundNotifyURL()
	body.Amount.Refund = in.RefundFen
	body.Amount.Total = in.TotalFen
	body.Amount.Currency = "CNY"
	var out Refund
	if err := c.do(ctx, http.MethodPost, "/v3/refund/domestic/refunds", body, &out); err != nil {
		return nil, err
	}
	if out.RefundID == "" {
		return nil, errors.New("wechatpay: 退款受理成功但未返回 refund_id")
	}
	return &out, nil
}

// QueryRefund 主动查退款单——与查支付单同理，**这是唯一可信的退款到账依据**。
func (c *Client) QueryRefund(ctx context.Context, outRefundNo string) (*Refund, error) {
	if outRefundNo == "" {
		return nil, errors.New("wechatpay: 退款单号必填")
	}
	var out Refund
	if err := c.do(ctx, http.MethodGet, "/v3/refund/domestic/refunds/"+url.PathEscape(outRefundNo), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
