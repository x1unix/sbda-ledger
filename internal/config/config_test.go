package config

import (
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFromFile(t *testing.T) {
	cases := []struct {
		label     string
		srcFile   string
		envs      map[string]string
		expect    *Config
		expectErr string
	}{
		{
			label:     "handle bad file path",
			srcFile:   "not-existing-file-name",
			expectErr: `failed to open config file: open testdata/not-existing-file-name: no such file or directory`,
		},
		{
			label:     "handle invalid YAML",
			srcFile:   "bad.yml",
			expectErr: `failed to parse config file "testdata/bad.yml"`,
		},
		{
			label:   "load full config",
			srcFile: "cfg.full.yml",
			expect: &Config{
				Listen: ":1",
				DB: Database{
					Address:             "fulladdr",
					MigrationsDirectory: "dir",
				},
				Redis: Redis{
					Address:  "redisaddr",
					DB:       1111,
					Password: "pass",
				},
			},
		},
		{
			label:   "fill empty with defaults",
			srcFile: "cfg.empty.yml",
			expect: &Config{
				Listen: ":88",
				DB: Database{
					Address:             "postgres://localhost:5432/ledger",
					MigrationsDirectory: "db/migrations",
				},
				Redis: Redis{
					Address: "localhost:6379",
				},
			},
		},
		{
			label:   "fill empty values with env vars and defaults",
			srcFile: "cfg.min.yml",
			expect: &Config{
				Listen: ":8888",
				DB: Database{
					Address:             "fulladdr",
					MigrationsDirectory: "testmigrations",
				},
				Redis: Redis{
					DB:       1234,
					Address:  "localhost:6379",
					Password: "redispass",
				},
			},
			envs: map[string]string{
				envPrefixed("REDIS_DB"):       "1234",
				envPrefixed("REDIS_PASSWORD"): "redispass",
				envPrefixed("MIGRATIONS_DIR"): "testmigrations",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.label, useEnvVars(c.envs, func(t *testing.T) {
			cfg, err := FromFile(filepath.Join("testdata", c.srcFile))
			if c.expectErr != "" {
				require.Error(t, err)
				if !strings.Contains(err.Error(), c.expectErr) {
					t.Fatalf("error '%s' should contain '%s'", err, c.expectErr)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, cfg)
			require.Equal(t, *c.expect, *cfg)
		}))
	}
}

func TestFromEnv(t *testing.T) {
	cases := []struct {
		label string
		envs  map[string]string
		want  Config
	}{
		{
			label: "valid default config",
			want: Config{
				Listen: ":8080",
				DB: Database{
					Address:             "postgres://localhost:5432/ledger",
					MigrationsDirectory: "db/migrations",
				},
				Redis: Redis{
					Address: "localhost:6379",
				},
			},
		},
		{
			label: "valid config from environment",
			want: Config{
				Listen: ":10541",
				DB: Database{
					Address:             "postgres://pguser:pgpass@localhost:5432/ledger",
					MigrationsDirectory: "/tmp/migrations",
				},
				Redis: Redis{
					DB:       14,
					Address:  "localhost:16379",
					Password: "fisheye",
				},
			},
			envs: map[string]string{
				envPrefixed("HTTP_ADDR"):      ":10541",
				envPrefixed("DB_ADDRESS"):     "postgres://pguser:pgpass@localhost:5432/ledger",
				envPrefixed("MIGRATIONS_DIR"): "/tmp/migrations",
				envPrefixed("REDIS_DB"):       "14",
				envPrefixed("REDIS_ADDRESS"):  "localhost:16379",
				envPrefixed("REDIS_PASSWORD"): "fisheye",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.label, useEnvVars(c.envs, func(t *testing.T) {
			got, err := FromEnv()
			require.NoError(t, err)
			require.NotNil(t, got)
			require.Equal(t, c.want, *got)
		}))
	}
}

func envPrefixed(name string) string {
	return "LGR_" + name
}

func useEnvVars(vars map[string]string, fn func(t *testing.T)) func(t *testing.T) {
	return func(t *testing.T) {
		if len(vars) == 0 {
			fn(t)
			return
		}

		for key, val := range vars {
			t.Logf("Env: set %q = '%s'", key, val)
			os.Setenv(key, val)
		}

		defer func() {
			var err error
			for k := range vars {
				t.Logf("Env: unset %q", k)
				if err = os.Unsetenv(k); err != nil {
					t.Logf("WARN: failed to unset %q - %v", k, err)
				}
			}
		}()

		fn(t)
	}
}
