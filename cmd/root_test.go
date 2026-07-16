package cmd

import (
	"strconv"
	"testing"

	"github.com/longkey1/gdoc/internal/gdoc"
	"github.com/spf13/cobra"
)

func newTestCmd(write bool) *cobra.Command {
	c := &cobra.Command{Use: "test"}
	if write {
		c.Annotations = map[string]string{writeAnnotation: "true"}
	}
	c.Flags().Bool("read-only", false, "")
	return c
}

// checkReadOnly reads the package-level config and readOnlyFlag globals, so
// these subtests reset them and must not run in parallel.
func TestCheckReadOnly(t *testing.T) {
	origConfig, origFlag := config, readOnlyFlag
	t.Cleanup(func() {
		config = origConfig
		readOnlyFlag = origFlag
	})

	tests := []struct {
		name      string
		write     bool
		cfg       *gdoc.Config
		flagValue *bool // nil means the flag was not passed on the command line
		wantErr   bool
	}{
		{
			name:    "non-write command always allowed even when config is read-only",
			write:   false,
			cfg:     &gdoc.Config{ReadOnly: true},
			wantErr: false,
		},
		{
			name:    "write command allowed when config is not read-only",
			write:   true,
			cfg:     &gdoc.Config{ReadOnly: false},
			wantErr: false,
		},
		{
			name:    "write command blocked when config is read-only",
			write:   true,
			cfg:     &gdoc.Config{ReadOnly: true},
			wantErr: true,
		},
		{
			name:    "write command allowed when no config is loaded",
			write:   true,
			cfg:     nil,
			wantErr: false,
		},
		{
			name:      "explicit --read-only=false overrides a read-only config",
			write:     true,
			cfg:       &gdoc.Config{ReadOnly: true},
			flagValue: new(false),
			wantErr:   false,
		},
		{
			name:      "explicit --read-only=true overrides a writable config",
			write:     true,
			cfg:       &gdoc.Config{ReadOnly: false},
			flagValue: new(true),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config = tt.cfg
			readOnlyFlag = false

			cmd := newTestCmd(tt.write)
			if tt.flagValue != nil {
				readOnlyFlag = *tt.flagValue
				if err := cmd.Flags().Set("read-only", strconv.FormatBool(*tt.flagValue)); err != nil {
					t.Fatal(err)
				}
			}

			err := checkReadOnly(cmd, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkReadOnly() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
