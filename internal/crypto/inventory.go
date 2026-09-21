package crypto

// Column 标识一个**应当是 AES-256-GCM 密文**的列。
type Column struct {
	Table  string
	Column string
}

// EncryptedColumns 是全库**唯一一份**加密列清单。
//
// 为什么集中放在这里（2026-09-21 彻查发现）：轮换工具 cmd/reencrypt 与校验工具
// cmd/keycheck 此前各带一份清单，而且两份都不全 —— reencrypt 漏了
// competition_registrations.id_card / .phone，而那两个列**确实是密文**（仓储层
// compRepo.encRegPII 加密，生产实测值形如 AW1StJscYi6K5oDgN/rniKYKzYMaO7）。
// 后果是：一旦用 reencrypt 轮换 ENCRYPTION_KEY，赛事报名的身份证与手机号会被
// 当成"历史明文"跳过，轮换后旧密钥丢弃 → 那些数据**永久解不开**，而
// decRegPII 解不开时会保留原值，界面就直接把密文当身份证显示。
//
// 清单漂移的代价是静默的数据损坏，所以只留这一份，两个工具都 import 它；
// 新增加密列时改这里一处即可。
var EncryptedColumns = []Column{
	{"demands", "contact"},
	{"enterprises", "license_url"},
	{"enterprises", "account_name"},
	{"users", "phone_ciphertext"},
	{"certified_pilots", "id_card"},
	{"competition_registrations", "id_card"},
	{"competition_registrations", "phone"},
	// 2026-09-21 补：培训报名的实名信息此前是**明文入库**（实测 id_card=500202100766642255、
	// phone=19823864146），而飞手档案与赛事报名早已加密 —— 同一个仓、同一把 cipher，只有这条
	// 链路漏了。写路径见 phase3_repos2.go 的 enrollRepo.encPII / decPII。
	{"training_enrollments", "id_card"},
	{"training_enrollments", "phone"},
}
