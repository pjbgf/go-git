package hash

// import (
// 	"testing"
// )

// func newI() ImmutableHash {
// 	return SHA1Hash{}
// }

// func newSHA1() SHA1Hash {
// 	return SHA1Hash{}
// }

// func newSHA256() SHA256Hash {
// 	return SHA256Hash{}
// }

// func BenchmarkZero(b *testing.B) {
// 	// var h StringHash

// 	b.Run("sha1", func(b *testing.B) {
// 		for i := 0; i < b.N; i++ {
// 			_ = newSHA1()
// 		}
// 	})
// 	b.Run("sha256", func(b *testing.B) {
// 		for i := 0; i < b.N; i++ {
// 			_ = newSHA256()
// 		}
// 	})
// 	b.Run("interface", func(b *testing.B) {
// 		for i := 0; i < b.N; i++ {
// 			_ = newI()
// 		}
// 	})
// }
