package eth

import (
	"testing"
)

func TestRpcChecker(t *testing.T) {
	t.Skip()
	checker := NewRpcChecker([]string{})
	checker.GetExtraRpcs(80001)
}
