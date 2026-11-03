package kalign_test

import (
	"bytes"
	"errors"
	"testing"

	"terraform-provider-kion/internal/kalign"
	"terraform-provider-kion/internal/kalign/mocks"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func testOpts() kalign.Options {
	return kalign.Options{SDKDir: "sdk", Version: "v3_16", ServiceRoot: "svc", FlexDir: "flex"}
}

func TestOptions_SDKFile(t *testing.T) {
	o := kalign.Options{SDKDir: "a/b", Version: "v3_16"}
	require.Equal(t, "a/b/generated/v3_16/oas_schemas_gen.go", o.SDKFile())
}

func TestCheck_WithMockSource(t *testing.T) {
	m := mocks.NewMockSource(t)
	o := testOpts()
	m.EXPECT().SDKStructs(o.SDKFile()).Return(map[string][]kalign.SDKField{
		"OUNote": {{GoName: "ID", JSON: "id", GoType: "OptUint64"}},
	}, nil)
	m.EXPECT().FlexFuncs("flex").Return(map[string]bool{"OptUint64ToFramework": true}, nil)
	m.EXPECT().ServiceModels("svc", "").Return([]kalign.ServiceModel{{
		Service: "ou_note", Name: "OuNoteModel",
		Fields: []kalign.ModelField{
			{GoName: "Id", TFSDK: "id", TFType: "types.Int64"},
			{GoName: "Ghost", TFSDK: "ghost", TFType: "types.String"}, // drift
		},
	}}, nil)

	var b bytes.Buffer
	findings, err := kalign.Check(m, &b, o)
	require.NoError(t, err)
	// The only finding here is a DRIFT one, which is advisory: it is printed
	// but not counted, so the run reports nothing to fix. See
	// Resolved.Actionable.
	require.Empty(t, findings)
	require.Contains(t, b.String(), "DRIFT no SDK field")
	require.Contains(t, b.String(), "(advisory)")
}

func TestGen_WithMockSource(t *testing.T) {
	m := mocks.NewMockSource(t)
	m.EXPECT().SDKStructs(mock.Anything).Return(map[string][]kalign.SDKField{
		"OUNote": {{GoName: "ID", JSON: "id", GoType: "OptUint64"}},
	}, nil)
	m.EXPECT().FlexFuncs(mock.Anything).Return(map[string]bool{"OptUint64ToFramework": true}, nil)
	m.EXPECT().ServiceModels(mock.Anything, mock.Anything).Return([]kalign.ServiceModel{{
		Service: "ou_note", Name: "OuNoteModel",
		Fields: []kalign.ModelField{{GoName: "Id", TFSDK: "id", TFType: "types.Int64"}},
	}}, nil)

	var b bytes.Buffer
	todos, err := kalign.Gen(m, &b, testOpts())
	require.NoError(t, err)
	require.Equal(t, 0, todos)
	require.Contains(t, b.String(), "m.Id = flex.OptUint64ToFramework(in.ID)")
}

func TestGen_ErrorPropagates(t *testing.T) {
	m := mocks.NewMockSource(t)
	m.EXPECT().SDKStructs(mock.Anything).Return(nil, errors.New("boom"))
	_, err := kalign.Gen(m, &bytes.Buffer{}, testOpts())
	require.Error(t, err)
}

func TestCheck_ErrorsPropagate(t *testing.T) {
	sentinel := errors.New("boom")

	t.Run("sdk", func(t *testing.T) {
		m := mocks.NewMockSource(t)
		m.EXPECT().SDKStructs(mock.Anything).Return(nil, sentinel)
		_, err := kalign.Check(m, &bytes.Buffer{}, testOpts())
		require.ErrorIs(t, err, sentinel)
	})
	t.Run("flex", func(t *testing.T) {
		m := mocks.NewMockSource(t)
		m.EXPECT().SDKStructs(mock.Anything).Return(map[string][]kalign.SDKField{}, nil)
		m.EXPECT().FlexFuncs(mock.Anything).Return(nil, sentinel)
		_, err := kalign.Check(m, &bytes.Buffer{}, testOpts())
		require.ErrorIs(t, err, sentinel)
	})
	t.Run("models", func(t *testing.T) {
		m := mocks.NewMockSource(t)
		m.EXPECT().SDKStructs(mock.Anything).Return(map[string][]kalign.SDKField{}, nil)
		m.EXPECT().FlexFuncs(mock.Anything).Return(map[string]bool{}, nil)
		m.EXPECT().ServiceModels(mock.Anything, mock.Anything).Return(nil, sentinel)
		_, err := kalign.Check(m, &bytes.Buffer{}, testOpts())
		require.ErrorIs(t, err, sentinel)
	})
}
