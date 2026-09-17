//go:build !race

package httpapi_test

// raceEnabled 见 race_on_test.go：非 race 构建下才断言吞吐阈值。
const raceEnabled = false
