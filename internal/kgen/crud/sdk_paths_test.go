package crud

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	// A call through any receiver: conn.GetX(, client.GetX(, d.Meta().Client.GetX(.
	templateCall = regexp.MustCompile(`\.([A-Z][A-Za-z0-9]*)\(`)
	// A by-path reference: (.Op "GET /v3/billing-rule/{id}").
	templateOpRef = regexp.MustCompile(`\.Op "([A-Z]+ [^"]+)"`)
	providerSDK   = regexp.MustCompile(`kion-sdk-go/generated/(v\d+_\d+)"`)
)

// currentSDKVersion is the sub-package the provider compiles against, read from
// its client wrapper (versions.CurrentSDK would be an import cycle).
func currentSDKVersion(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "conns", "kion_client.go"))
	require.NoError(t, err)
	m := providerSDK.FindSubmatch(b)
	require.NotNil(t, m, "internal/conns/kion_client.go names no SDK sub-package")
	return string(m[1])
}

func sdkModuleDir(t *testing.T) string {
	t.Helper()
	out, err := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/kionsoftware/kion-sdk-go").Output()
	require.NoError(t, err, "locating kion-sdk-go")
	dir := strings.TrimSpace(string(out))
	require.NotEmpty(t, dir, "kion-sdk-go is not downloaded (go mod download)")
	return dir
}

// Every SDK method a verbatim template names must still be the route it was
// written for. ogen assigns names by spec order, so a bump can hand a name to a
// different route (v3_17 gave GetBillingRule to /beta) and the result compiles.
func TestTemplateSDKCallsResolveToIntendedPaths(t *testing.T) {
	idx, err := buildIndex(NewFileSource(), sdkModuleDir(t), currentSDKVersion(t))
	require.NoError(t, err)
	byName := map[string]ClientMethod{}
	for _, m := range idx.methods {
		byName[m.Name] = m
	}

	seen := map[string]map[string]bool{}
	require.NoError(t, filepath.WalkDir(".", func(p string, d fs.DirEntry, err error) error {
		require.NoError(t, err)
		if d.IsDir() || !strings.HasSuffix(p, ".gtpl") {
			return nil
		}
		b, rerr := os.ReadFile(p)
		require.NoError(t, rerr)
		for _, m := range templateCall.FindAllSubmatch(b, -1) {
			name := string(m[1])
			got, isSDK := byName[name]
			if !isSDK {
				continue
			}
			want, declared := templateSDKPaths[p][name]
			if !declared {
				t.Errorf("%s calls SDK method %s (%s %s) with no intended path in templateSDKPaths; resolve it by path with .Op, or declare it", p, name, got.HTTPMethod, got.Path)
				continue
			}
			if seen[p] == nil {
				seen[p] = map[string]bool{}
			}
			seen[p][name] = true
			require.Equalf(t, want, got.HTTPMethod+" "+got.Path,
				"%s calls %s for %s, but in this SDK that name is %s %s", p, name, want, got.HTTPMethod, got.Path)
		}
		for _, m := range templateOpRef.FindAllSubmatch(b, -1) {
			_, ok := idx.methods[string(m[1])]
			require.Truef(t, ok, "%s resolves %q by path, and this SDK has no such route", p, m[1])
		}
		return nil
	}))

	for file, calls := range templateSDKPaths {
		for name := range calls {
			require.Truef(t, seen[file][name], "templateSDKPaths declares %s in %s, which no longer calls it", name, file)
		}
	}
}
