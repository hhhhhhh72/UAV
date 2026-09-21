// keycheck 校验「备份 + 当前 ENCRYPTION_KEY」能否读出加密字段。
//
// 为什么需要（2026-09-21 彻查）：备份里躺着 PII 密文，而密钥只存在于本机的
// /root/UAV/.env。备份再完好，只要密钥丢了或被轮换过，那些身份证/手机号就**永久
// 解不开** —— 而在此之前没有任何东西会发现这件事：备份校验过 gzip、演练数过表数，
// 全都「通过」。这个工具把「密钥与备份是否配套」变成一条可验证的结论。
//
// 它**复用 internal/crypto 的同一份实现**，不在演练脚本里另写一套解密逻辑 ——
// 两套实现迟早漂移，那时演练通过与否就与真实解密能力无关了。
//
// 判定分三类（避免把「明文入库」误判成「密钥不对」）：
//
//	ok        能解开 → 格式与密钥都对
//	plaintext 值本身像明文（身份证/手机号格式）→ 记为明文行，不算失败
//	bad       既解不开、又不像明文 → **失败**（密钥漂移/密文损坏）
//
// 用法：keycheck <database_url>
//
//	keycheck "postgres://drone@/restore_drill?host=/var/run/postgresql"
//
// 环境：ENCRYPTION_KEY（base64 的 32 字节，与 API 同一把）
package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"drone-platform/internal/crypto"
)

// 加密列清单只有一份：internal/crypto.EncryptedColumns（见该文件注释）。
// 这里**不再各写一份** —— 两份不一致正是轮换工具漏掉 competition_registrations
// 的原因。

// looksLikeCiphertext 判断一个值**有没有可能是** base64 密文。
// 这是「解不开」与「本来就是明文」的分界线：GCM 密文一定是合法 base64，
// 而中文企业名、18 位身份证号（长度 18 % 4 = 2）都不是 —— 解不开又不像 base64
// 的值应记为「明文行」而不是「密钥不对」。
//
// 这条规则是实测逼出来的：最初只认身份证/手机号格式，于是
// enterprises.account_name 里两行中文企业名被误报成「无法解密」，整个演练变红。
func looksLikeCiphertext(v string) bool {
	_, err := base64.StdEncoding.DecodeString(v)
	return err == nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: keycheck <database_url>")
		os.Exit(2)
	}
	key := os.Getenv("ENCRYPTION_KEY")
	if key == "" {
		fmt.Println("FAIL 未设置 ENCRYPTION_KEY —— 备份里的密文无法校验")
		os.Exit(2)
	}
	cph, err := crypto.NewCipher(key)
	if err != nil {
		fmt.Printf("FAIL 密钥无效（应为 base64 的 32 字节）: %v\n", err)
		os.Exit(2)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, os.Args[1])
	if err != nil {
		fmt.Printf("FAIL 连接数据库: %v\n", err)
		os.Exit(2)
	}
	defer pool.Close()

	var total, ok, plain, bad int
	var samples []string
	for _, ec := range crypto.EncryptedColumns {
		var cTotal, cOK, cPlain, cBad int
		rows, err := pool.Query(ctx,
			"SELECT "+ec.Column+"::text FROM "+ec.Table+" WHERE COALESCE("+ec.Column+",'') <> ''")
		if err != nil {
			fmt.Printf("WARN 读取 %s.%s 失败（表可能不存在）: %v\n", ec.Table, ec.Column, err)
			continue
		}
		for rows.Next() {
			var v string
			if err := rows.Scan(&v); err != nil {
				continue
			}
			total++
			cTotal++
			switch {
			case decryptable(cph, v):
				ok++
				cOK++
			case !looksLikeCiphertext(v):
				plain++
				cPlain++
			default:
				bad++
				cBad++
				if len(samples) < 3 {
					head := v
					if len(head) > 12 {
						head = head[:12]
					}
					samples = append(samples, ec.Table+"."+ec.Column+"="+head+"...")
				}
			}
		}
		rows.Close()
		if cTotal > 0 {
			fmt.Printf("  %-42s 共 %d：密文可解 %d，明文 %d，**解不开 %d**\n", ec.Table+"."+ec.Column, cTotal, cOK, cPlain, cBad)
		}
	}

	fmt.Printf("密文行 %d：可解密 %d，疑似明文 %d，无法解密 %d\n", total, ok, plain, bad)
	for _, s := range samples {
		fmt.Println("  失败样本:", s)
	}
	switch {
	case bad > 0:
		fmt.Println("FAIL 存在无法解密的密文 —— 密钥与备份不配套（被轮换过？），恢复出来的 PII 将读不出来")
		os.Exit(1)
	case total == 0:
		fmt.Println("WARN 备份里没有可校验的密文样本 —— 这一步证明不了「密钥与备份配套」，别当成通过")
		os.Exit(0)
	case ok == 0:
		fmt.Println("WARN 所有样本都像明文（PII 未加密入库）—— 密钥校验无意义")
		os.Exit(0)
	case plain > 0:
		fmt.Printf("OK 密文全部可解（%d 行）；另有 %d 行是明文 —— 不是密钥问题，但说明加密范围没覆盖到，建议单独核对\n", ok, plain)
		os.Exit(0)
	default:
		fmt.Printf("OK 加密字段可解密（%d 行），密钥与备份配套\n", ok)
	}
}

func decryptable(c *crypto.Cipher, v string) bool {
	out, err := c.Decrypt(v)
	return err == nil && out != ""
}
