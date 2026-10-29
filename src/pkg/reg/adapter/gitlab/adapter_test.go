package gitlab

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	adp "github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

func mustWriteHTTPResponse(t *testing.T, w io.Writer, fixturePath string) {
	f, err := os.Open(fixturePath)
	if err != nil {
		t.Fatalf("error opening fixture file: %v", err)
	}

	if _, err = io.Copy(w, f); err != nil {
		t.Fatalf("error writing response: %v", err)
	}
}
func getServer(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/v2/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Www-Authenticate", "Bearer realm=\"http://"+r.Host+"/jwt/auth\",service=\"container_registry\"")
		w.WriteHeader(http.StatusUnauthorized)

	})
	mux.HandleFunc("/api/v4/projects", func(w http.ResponseWriter, r *http.Request) {
		search := r.URL.Query().Get("search")
		w.Header().Set("X-Next-Page", "")

		switch search {
		case "library/dev-docker", "library", "library/", "dev-docker/", "dev-docker":
			mustWriteHTTPResponse(t, w, "testdata/projects/dev-docker.json")
			break
		case "", "library/dockers":
			mustWriteHTTPResponse(t, w, "testdata/projects/all.json")
			break
		default:
			w.Header().Set("X-Next-Page", "")
			w.Write([]byte(`[]`))
			break
		}

	})
	for projectID := 1; projectID <= 5; projectID++ {
		mux.HandleFunc("/api/v4/projects/"+strconv.Itoa(projectID)+"/registry/repositories", func(w http.ResponseWriter, r *http.Request) {

			w.Header().Set("X-Next-Page", "")
			re := regexp.MustCompile(`projects/(?P<id>\d+)/registry`)
			match := re.FindStringSubmatch(r.RequestURI)
			mustWriteHTTPResponse(t, w, "testdata/repositories/"+match[1]+".json")

		})
		for repositoryID := 1; repositoryID <= 5; repositoryID++ {
			mux.HandleFunc("/api/v4/projects/"+strconv.Itoa(projectID)+"/registry/repositories/"+strconv.Itoa(repositoryID)+"1/tags", func(w http.ResponseWriter, r *http.Request) {

				w.Header().Set("X-Next-Page", "")
				re := regexp.MustCompile(`repositories/(?P<id>\d+)/tags`)
				match := re.FindStringSubmatch(r.RequestURI)
				mustWriteHTTPResponse(t, w, "testdata/tags/"+match[1]+".json")

			})

		}
	}
	server := httptest.NewServer(mux)
	return server
}

func getAdapter(t *testing.T) adp.Adapter {
	assertions := assert.New(t)
	factory, err := adp.GetFactory(model.RegistryTypeGitLab)
	assertions.Nil(err)
	assertions.NotNil(factory)
	server := getServer(t)
	adapter, err := newAdapter(&model.Registry{
		Type: model.RegistryTypeGitLab,
		URL:  server.URL,
		Credential: &model.Credential{
			AccessKey:    "test",
			AccessSecret: "test",
		},
	})
	assertions.Nil(err)
	assertions.NotNil(adapter)

	return adapter
}
func TestFetchImages(t *testing.T) {
	assertions := assert.New(t)

	ad := getAdapter(t)
	adapter := ad.(*adapter)
	templates := map[string]int{
		"library/dev-docker": 1,
		"library/*-docker":   1,
		"library/dev-*":      2,
		"*/dev-docker":       1,
		"library/*":          2,
	}
	for k, v := range templates {
		resources, err := adapter.FetchArtifacts([]*model.Filter{
			{
				Type:  model.FilterTypeName,
				Value: k,
			},
		})
		require.Nil(t, err, k, v)
		assertions.Len(resources, v, k, v)
	}

	resources, err := adapter.FetchArtifacts([]*model.Filter{
		{
			Type:  model.FilterTypeName,
			Value: "library/dockers",
		},
		{
			Type:  model.FilterTypeTag,
			Value: "{late*,v2}",
		},
	})
	require.Nil(t, err)
	require.Equal(t, 1, len(resources))
	require.Equal(t, 2, len(resources[0].Metadata.Vtags))
}

func TestGitlabRegexFilters(t *testing.T) {
	client := getAdapter(t).(*adapter)
	resources, err := client.FetchArtifacts([]*model.Filter{
		{Type: model.FilterTypeName, Kind: "regex", Value: `library/(dockers)`},
		{Type: model.FilterTypeTag, Kind: "regex", Value: `latest|v2`},
	})
	require.NoError(t, err)
	require.Len(t, resources, 1)
	require.Len(t, resources[0].Metadata.Vtags, 2)
	resources, err = client.FetchArtifacts([]*model.Filter{
		{Type: model.FilterTypeName, Kind: "regex", Value: `LIBRARY/(dockers)`},
	})
	require.NoError(t, err)
	require.Empty(t, resources, "regex matching must be case-sensitive by default")
}

func TestGitlabRegexExclusionPreservesLegacyGlobBehavior(t *testing.T) {
	client := getAdapter(t).(*adapter)
	for _, tt := range []struct {
		kind string
		want []string
	}{
		{"", []string{"v2"}}, {"doublestar", []string{"v2"}}, {"regex", []string{"harbor", "latest"}},
	} {
		resources, err := client.FetchArtifacts([]*model.Filter{
			{Type: model.FilterTypeName, Value: "library/dockers"},
			{Type: model.FilterTypeTag, Kind: tt.kind, Value: "v2", Decoration: model.Excludes},
		})
		require.NoError(t, err)
		require.Len(t, resources, 1)
		require.Equal(t, tt.want, resources[0].Metadata.Vtags, tt.kind)
	}
}
