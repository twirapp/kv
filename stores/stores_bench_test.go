package stores

import (
	"context"
	"testing"
)

func BenchmarkGet(b *testing.B) {
	for _, impl := range implementations {
		store := impl.create()
		key := "test_key"
		value := "test_value"

		ctx := context.Background()

		if err := store.Set(ctx, key, value); err != nil {
			b.Fatalf("failed to set initial value: %v", err)
		}

		b.ResetTimer()

		b.Run(impl.name, func(b *testing.B) {
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					v := store.Get(ctx, key)
					casted, err := v.String()
					if err != nil {
						b.Fatalf("failed to get value: %v", err)
					}

					if casted != value {
						b.Fatalf("unexpected value: got %s, want %s", casted, value)
					}
				}
			})
		})
	}
}

//func BenchmarkSet(b *testing.B) {
//	for _, impl := range implementations {
//		if impl.name == "Redis" {
//			// Skip Redis benchmark for now
//			continue
//		}
//
//		store := impl.create()
//		key := "test_key"
//		value := "test_value"
//
//		b.ResetTimer()
//		b.Run(impl.name, func(b *testing.B) {
//			for i := 0; i < b.N; i++ {
//				if err := store.Set(nil, fmt.Sprintf("%s:%s:%d", impl.name, key, i), value); err != nil {
//					b.Fatalf("failed to set value: %v", err)
//				}
//			}
//		})
//	}
//}
