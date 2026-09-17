//go:build race

package httpapi_test

// raceEnabled 在 `go test -race` 构建下为 true。
//
// 为什么需要它：竞态检测器会把执行拖慢约一个数量级，而 CI runner 的算力又远低于
// 开发机，于是「功能完全正确」的并发压测会因为吞吐不达标而报假失败。
// 2026-09-17 实测：extreme_test 的 ops>=10000 断言在 -race 下只跑到 7586，
// 让 CI 连红近一个月（前面的步骤先失败，这一步的病因一直没被人看到）。
//
// 吞吐属于基准测试的范畴，不该作为正确性断言在 race 构建下执行；
// 正确性断言（错误率、数据一致性）在两种构建下都保留。
const raceEnabled = true
