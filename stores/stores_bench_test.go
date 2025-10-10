package stores

import "testing"

func BenchmarkGet(b *testing.B) {
	for _, impl := range implementations {
		store := impl.create()
		key := "test_key"
		value := "test_value"

		if err := store.Set(nil, key, value); err != nil {
			b.Fatalf("failed to set initial value: %v", err)
		}

		b.ResetTimer()
		b.Run(impl.name, func(b *testing.B) {

			for i := 0; i < b.N; i++ {
				v := store.Get(nil, key)
				casted, err := v.String()
				if err != nil {
					b.Fatalf("failed to get value: %v", err)
				}

				if casted != value {
					b.Fatalf("unexpected value: got %s, want %s", casted, value)
				}
			}
		})
	}
}
