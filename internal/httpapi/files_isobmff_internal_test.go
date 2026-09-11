package httpapi

import "testing"

// ISO BMFF 判定（mp4/mov）：首 4 字节 box 长度 + "ftyp"。
//
// 为什么要自己判：Go 的 http.DetectContentType 对 mp4 只认少数 brand（isom/mp41/avc1…），
// 国产剪辑工具导出的 mp4 会被判成 application/octet-stream——上传时被魔数白名单拒绝，
// 取回时类型头不对又会导致播放器拒播。这里按容器结构判定，两处都用它兜底。
func TestLooksLikeISOBMFF(t *testing.T) {
	ftyp := append([]byte{0x00, 0x00, 0x00, 0x18}, []byte("ftypisom")...)
	ftyp = append(ftyp, []byte{0x00, 0x00, 0x02, 0x00}...)
	ftyp = append(ftyp, []byte("qt  mp42")...) // 非常见 brand 也要认
	if !looksLikeISOBMFF(ftyp) {
		t.Fatal("带 ftyp box 的文件应判为 mp4/mov 容器")
	}
	if looksLikeISOBMFF([]byte("this is just text, not a video")) {
		t.Fatal("普通文本不应判为 mp4")
	}
	if looksLikeISOBMFF([]byte("ftyp")) {
		t.Fatal("过短的数据不应判为 mp4（需 >=12 字节）")
	}
	if looksLikeISOBMFF([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D}) {
		t.Fatal("PNG 不应判为 mp4")
	}
}
