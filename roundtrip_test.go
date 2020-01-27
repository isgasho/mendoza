package mendoza_test

import (
	"encoding/json"
	"fmt"
	"github.com/sanity-io/mendoza"
	"github.com/stretchr/testify/require"
	"testing"
)

var Documents = []struct {
	Left  string
	Right string
}{
	{
		`{}`,
		`{}`,
	},
	{
		`1`,
		`{}`,
	},
	{
		`{"a": "b"}`,
		`{"a": "b"}`,
	},
	{
		`{"a": "a"}`,
		`{"a": "b"}`,
	},
	{
		`{"a": "a", "b": "b"}`,
		`{"a": "b"}`,
	},
	{
		`{"a": "a", "b": "b", "c": "c"}`,
		`{"a": "a", "b": "b", "c": "c", "d": "d"}`,
	},
	{
		`{"a": "a", "b": "b", "c": "c"}`,
		`{"d": "d"}`,
	},
	{
		`{"a": "a", "b": {"a": "a"}}`,
		`{"a": "a", "b": {"a": "b", "b": "a"}}`,
	},
	{
		`{"a": ["a", "b", "c"]}`,
		`{"a": ["a", "b", "c"]}`,
	},
	{
		`{"a": ["a", "b", "c"]}`,
		`{"a": ["a", "b"]}`,
	},
	{
		`{"a": [1, 2]}`,
		`{"a": [2, 3]}`,
	},
}

func TestRoundtrip(t *testing.T) {
	for idx, pair := range Documents {
		t.Run(fmt.Sprintf("N%d", idx), func(t *testing.T) {
			var left, right interface{}

			err := json.Unmarshal([]byte(pair.Left), &left)
			require.NoError(t, err)

			err = json.Unmarshal([]byte(pair.Right), &right)
			require.NoError(t, err)

			patch1, patch2, err := mendoza.CreateDoublePatch(left, right)
			require.NoError(t, err)

			result1 := mendoza.ApplyPatch(left, patch1)
			require.EqualValues(t, right, result1)

			result2 := mendoza.ApplyPatch(right, patch2)
			require.EqualValues(t, left, result2)

			// Now try to encode and decode the patch
			json1, err := json.Marshal(patch1)
			require.NoError(t, err)
			var parsedPatch1 mendoza.Patch
			err = json.Unmarshal(json1, &parsedPatch1)
			require.NoError(t, err)
			require.EqualValues(t, patch1, parsedPatch1)

			json2, err := json.Marshal(patch2)
			require.NoError(t, err)
			var parsedPatch2 mendoza.Patch
			err = json.Unmarshal(json2, &parsedPatch2)
			require.NoError(t, err)
			require.EqualValues(t, patch2, parsedPatch2)
		})
	}
}