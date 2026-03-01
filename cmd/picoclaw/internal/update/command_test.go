package update

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUpdateCommand(t *testing.T) {
	cmd := NewUpdateCommand()

	require.NotNil(t, cmd)

	assert.Equal(t, "update", cmd.Use)

	assert.Equal(t, "Update picoclaw to the latest version", cmd.Short)

	assert.Len(t, cmd.Aliases, 0)

	assert.True(t, cmd.HasFlags())

	f := cmd.Flags().Lookup("check")
	require.NotNil(t, f)
	assert.Equal(t, "false", f.DefValue)
	assert.Equal(t, "c", f.Shorthand)

	assert.False(t, cmd.HasSubCommands())

	assert.Nil(t, cmd.Run)
	assert.NotNil(t, cmd.RunE)

	assert.Nil(t, cmd.PersistentPreRun)
	assert.Nil(t, cmd.PersistentPostRun)
}

func TestUpdateCommand_CheckFlag(t *testing.T) {
	cmd := NewUpdateCommand()

	require.NotNil(t, cmd)

	f := cmd.Flags().Lookup("check")
	require.NotNil(t, f)

	err := cmd.Flags().Set("check", "true")
	require.NoError(t, err)

	assert.Equal(t, "true", f.Value.String())
}
