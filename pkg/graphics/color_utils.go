// color_utils.go は色変換に関するユーティリティ関数を提供する
// any型からintへの変換など、色値の処理に使用される共通関数を含む
package graphics

// toIntFromAny converts any to int.
// Supports int, int64, float64, and float32 types.
func toIntFromAny(v any) (int, bool) {
	switch val := v.(type) {
	case int:
		return val, true
	case int64:
		return int(val), true
	case float64:
		return int(val), true
	case float32:
		return int(val), true
	default:
		return 0, false
	}
}
