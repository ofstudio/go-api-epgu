package aas

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPermissions_Base64String(t *testing.T) {
	p := Permissions{
		Permission{
			ResponsibleObject: "test",
			Sysname:           "test",
			Expire:            1,
			Actions:           []PermissionAction{{Sysname: "test"}},
			Purposes:          []PermissionPurpose{{Sysname: "test"}},
			Scopes:            []PermissionScope{{Sysname: "test"}},
		},
	}

	base64String := p.Base64String()
	jsonBytes, err := base64.RawURLEncoding.DecodeString(base64String)
	require.NoError(t, err)
	var pGot Permissions
	err = json.Unmarshal(jsonBytes, &pGot)
	require.NoError(t, err)
	require.Equal(t, p, pGot)
}
