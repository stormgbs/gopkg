package utils

import (
	"testing"
)

var camelStringCases = [][2]string{
	[2]string{"GaoBuShuang", "gao_bu_shuang"},
	[2]string{"gaoBuShuang", "gao_bu_shuang"},
	[2]string{"Gao_buShuang", "gao_bu_shuang"},
	[2]string{"Gao_BuShuang", "gao_bu_shuang"},
}

func TestSnakeString(t *testing.T) {
	for _, sample := range camelStringCases {
		dest := SnakeString(sample[0])

		if dest != sample[1] {
			t.Errorf("Source: %s, got: %s, expect: %s", sample[0], dest, sample[1])
		}
	}
}
