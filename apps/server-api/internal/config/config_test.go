package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateForRuntimeRequiresPersistentProductionConfiguration(t *testing.T) {
	t.Setenv("GIN_MODE", "release")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("SB3_STORAGE_DIR", "")
	t.Setenv("CORS_ALLOWED_ORIGINS", "")

	cfg := FromEnv()

	err := cfg.ValidateForRuntime()
	require.Error(t, err)
	require.ErrorContains(t, err, "DATABASE_URL")
	require.ErrorContains(t, err, "SB3_STORAGE_DIR")
	require.ErrorContains(t, err, "CORS_ALLOWED_ORIGINS")
}

func TestValidateForRuntimeAllowsLocalDefaultsOutsideReleaseMode(t *testing.T) {
	t.Setenv("GIN_MODE", "debug")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("SB3_STORAGE_DIR", "")
	t.Setenv("CORS_ALLOWED_ORIGINS", "")

	cfg := FromEnv()

	require.NoError(t, cfg.ValidateForRuntime())
}
