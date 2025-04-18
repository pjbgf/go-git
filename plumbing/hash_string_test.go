package plumbing

import "testing"

func BenchmarkZero(b *testing.B) {
	var h StringHash

	// b.Run("demand", func(b *testing.B) {
	// 	for i := 0; i < b.N; i++ {
	// 		h.StringDemand()
	// 	}
	// })
	// b.Run("const", func(b *testing.B) {
	// 	for i := 0; i < b.N; i++ {
	// 		h.StringConst()
	// 	}
	// })
	b.Run("string", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = h.String()
		}
	})
}
