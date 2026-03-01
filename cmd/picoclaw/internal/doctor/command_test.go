package doctor

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDoctorCommand(t *testing.T) {
	cmd := NewDoctorCommand()

	require.NotNil(t, cmd)

	assert.Equal(t, "doctor", cmd.Use)

	assert.Equal(t, "Diagnose common problems", cmd.Short)

	assert.Len(t, cmd.Aliases, 1)
	assert.True(t, cmd.HasAlias("d"))

	assert.True(t, cmd.HasFlags())

	f := cmd.Flags().Lookup("fix")
	require.NotNil(t, f)
	assert.Equal(t, "false", f.DefValue)

	assert.False(t, cmd.HasSubCommands())

	assert.Nil(t, cmd.Run)
	assert.NotNil(t, cmd.RunE)

	assert.Nil(t, cmd.PersistentPreRun)
	assert.Nil(t, cmd.PersistentPostRun)
}

func TestDoctorCommand_FixFlag(t *testing.T) {
	cmd := NewDoctorCommand()

	require.NotNil(t, cmd)

	f := cmd.Flags().Lookup("fix")
	require.NotNil(t, f)

	err := cmd.Flags().Set("fix", "true")
	require.NoError(t, err)

	assert.Equal(t, "true", f.Value.String())
}
