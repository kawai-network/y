package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigFromRPCURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		rpcURL        string
		wantEnv       Environment
		wantChainID   uint64
		wantIsTestnet bool
		wantErr       bool
	}{
		{
			name:          "testnet url",
			rpcURL:        "https://rpc.testnet.monad.xyz",
			wantEnv:       EnvironmentTestnet,
			wantChainID:   10143,
			wantIsTestnet: true,
		},
		{
			name:          "mainnet url",
			rpcURL:        "https://rpc.mainnet.monad.xyz",
			wantEnv:       EnvironmentMainnet,
			wantChainID:   143,
			wantIsTestnet: false,
		},
		{
			name:          "default monad rpc url",
			rpcURL:        "https://rpc.monad.xyz",
			wantEnv:       EnvironmentMainnet,
			wantChainID:   143,
			wantIsTestnet: false,
		},
		{
			name:    "invalid url",
			rpcURL:  "https://rpc.local",
			wantErr: true,
		},
		{
			name:    "empty url",
			rpcURL:  "",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg, err := configFromRPCURL(tc.rpcURL)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, cfg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, cfg)
			require.Equal(t, tc.wantEnv, cfg.Environment)
			require.Equal(t, tc.wantChainID, cfg.ChainID)
			require.Equal(t, tc.wantIsTestnet, cfg.IsTestnet)
			require.Equal(t, !tc.wantIsTestnet, cfg.IsMainnet)
		})
	}
}
