package sock_test

import (
	"context"
	"testing"

	"github.com/AR1011/wazero/experimental/sock"
	internalsock "github.com/AR1011/wazero/internal/sock"
	"github.com/AR1011/wazero/internal/testing/require"
)

// testCtx is an arbitrary, non-default context. Non-nil also prevents linter errors.
var testCtx = context.WithValue(context.Background(), struct{}{}, "arbitrary")

func TestWithSockConfig(t *testing.T) {
	tests := []struct {
		name     string
		sockCfg  sock.Config
		expected bool
	}{
		{
			name:     "returns input when sockCfg nil",
			expected: false,
		},
		{
			name:     "returns input when sockCfg empty",
			sockCfg:  sock.NewConfig(),
			expected: false,
		},
		{
			name:     "decorates with sockCfg",
			sockCfg:  sock.NewConfig().WithTCPListener("", 0),
			expected: true,
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			if decorated := sock.WithConfig(testCtx, tc.sockCfg); tc.expected {
				require.NotNil(t, decorated.Value(internalsock.ConfigKey{}))
			} else {
				require.Same(t, testCtx, decorated)
			}
		})
	}
}
