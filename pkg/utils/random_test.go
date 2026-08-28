package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateBase62Key_Length(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{name: "length 7 (default use case)", length: 7},
		{name: "length 1", length: 1},
		{name: "length 20", length: 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := GenerateBase62Key(tt.length)

			require.NoError(t, err)
			require.Len(t, key, tt.length)
		})
	}
}

func TestGenerateBase62Key_Charset(t *testing.T) {
	key, err := GenerateBase62Key(1000) // key dài để tăng khả năng cover hết charset
	require.NoError(t, err)

	for _, c := range key {
		require.True(t,
			strings.ContainsRune(base62Alphabet, c),
			"character %q is not in base62 alphabet", c,
		)
	}
}

func TestGenerateBase62Key_NoCollision(t *testing.T) {
	const numKeys = 100_000
	const keyLength = 7

	seen := make(map[string]struct{}, numKeys)

	for i := 0; i < numKeys; i++ {
		key, err := GenerateBase62Key(keyLength)
		require.NoError(t, err)

		_, exists := seen[key]
		require.False(t, exists, "collision detected at key: %s", key)

		seen[key] = struct{}{}
	}
}

func TestGenerateBase62Key_ZeroLength(t *testing.T) {
	key, err := GenerateBase62Key(0)

	require.NoError(t, err)
	require.Equal(t, "", key)
}

func BenchmarkGenerateBase62Key(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = GenerateBase62Key(7)
	}
}
