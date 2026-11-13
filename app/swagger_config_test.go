package app_test

import (
	"encoding/json"
	"os"
	"regexp"
	"slices"
	"testing"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// The server registers these services at start, not app.New. RegisterAPIRoutes
// serves their REST routes.
var serverQueryServices = []string{
	"cosmos.base.node.v1beta1.Service",
	"cosmos.base.tendermint.v1beta1.Service",
	"cosmos.tx.v1beta1.Service",
}

// These services have REST annotations, but nothing registers their gateway
// handler, so the app does not serve them. Baseapp registers the reflection
// service, and ibc-go registers the IBC client v2 service.
var queryServicesWithoutGateway = []string{
	"cosmos.base.reflection.v1beta1.ReflectionService",
	"ibc.core.client.v2.Query",
}

var swaggerURL = regexp.MustCompile(`^\./tmp-swagger-gen/(.+)\.swagger\.json$`)

// docs/swagger-config.json must list exactly the query services with REST
// routes that the app serves. Then make proto-swagger-gen documents a new
// module, and drops a removed one.
func TestSwaggerConfig_MatchesServedQueryServices(t *testing.T) {
	server := grpc.NewServer()
	sharedTestApp().RegisterGRPCServer(server)

	served := slices.Clone(serverQueryServices)
	for name := range server.GetServiceInfo() {
		if hasRESTRoutes(t, name) && !slices.Contains(queryServicesWithoutGateway, name) {
			served = append(served, name)
		}
	}

	raw, err := os.ReadFile("../docs/swagger-config.json")
	require.NoError(t, err)
	var config struct {
		APIs []struct {
			URL string `json:"url"`
		} `json:"apis"`
	}
	require.NoError(t, json.Unmarshal(raw, &config))

	var documented []string
	for _, api := range config.APIs {
		m := swaggerURL.FindStringSubmatch(api.URL)
		require.NotNil(t, m, api.URL)
		file, err := gogoproto.HybridResolver.FindFileByPath(m[1] + ".proto")
		require.NoError(t, err, api.URL)
		for i := range file.Services().Len() {
			documented = append(documented, string(file.Services().Get(i).FullName()))
		}
	}

	require.ElementsMatch(t, served, documented,
		"update docs/swagger-config.json, or add the service to queryServicesWithoutGateway if nothing registers its gateway handler")
}

func hasRESTRoutes(t *testing.T, service string) bool {
	t.Helper()
	desc, err := gogoproto.HybridResolver.FindDescriptorByName(protoreflect.FullName(service))
	require.NoError(t, err, service)
	sd, ok := desc.(protoreflect.ServiceDescriptor)
	require.True(t, ok, service)
	for i := range sd.Methods().Len() {
		if proto.HasExtension(sd.Methods().Get(i).Options(), annotations.E_Http) {
			return true
		}
	}
	return false
}
