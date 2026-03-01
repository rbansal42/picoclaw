package sessions

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSessionsCommand(t *testing.T) {
	cmd := NewSessionsCommand()

	require.NotNil(t, cmd)

	assert.Equal(t, "sessions", cmd.Use)

	assert.Equal(t, "Manage sessions", cmd.Short)

	assert.Len(t, cmd.Aliases, 0)

	assert.False(t, cmd.HasFlags())

	assert.Nil(t, cmd.Run)
	assert.NotNil(t, cmd.RunE)

	assert.NotNil(t, cmd.PersistentPreRunE)
	assert.Nil(t, cmd.PersistentPreRun)
	assert.Nil(t, cmd.PersistentPostRun)

	assert.True(t, cmd.HasSubCommands())

	allowedCommands := []string{
		"list",
		"show",
		"delete",
		"clear",
	}

	subcommands := cmd.Commands()
	assert.Len(t, subcommands, len(allowedCommands))

	for _, subcmd := range subcommands {
		found := slices.Contains(allowedCommands, subcmd.Name())
		assert.True(t, found, "unexpected subcommand %q", subcmd.Name())

		assert.Len(t, subcmd.Aliases, 0)
		assert.False(t, subcmd.Hidden)

		assert.False(t, subcmd.HasSubCommands())

		assert.Nil(t, subcmd.Run)
		assert.NotNil(t, subcmd.RunE)

		assert.Nil(t, subcmd.PersistentPreRun)
		assert.Nil(t, subcmd.PersistentPostRun)
	}
}

func TestSessionsSubcommands_ArgsValidation(t *testing.T) {
	cmd := NewSessionsCommand()

	require.NotNil(t, cmd)

	tests := []struct {
		name      string
		exactArgs bool
	}{
		{name: "show", exactArgs: true},
		{name: "delete", exactArgs: true},
		{name: "list", exactArgs: false},
		{name: "clear", exactArgs: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subcmd, _, err := cmd.Find([]string{tt.name})
			require.NoError(t, err)
			require.NotNil(t, subcmd)
			assert.Equal(t, tt.name, subcmd.Name())

			if tt.exactArgs {
				// Commands requiring exactly 1 arg should reject 0 args
				require.NotNil(t, subcmd.Args)
				err := subcmd.Args(subcmd, []string{})
				assert.Error(t, err)
				err = subcmd.Args(subcmd, []string{"test-id"})
				assert.NoError(t, err)
			} else {
				// Commands accepting no args should reject 1 arg
				require.NotNil(t, subcmd.Args)
				err := subcmd.Args(subcmd, []string{})
				assert.NoError(t, err)
				err = subcmd.Args(subcmd, []string{"unexpected"})
				assert.Error(t, err)
			}
		})
	}
}
