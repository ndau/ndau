package sdk_test

import (
	"net/http/httptest"
	"testing"

	sdk "github.com/oneiro-ndev/ndau/pkg/api_sdk"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/mock"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/svc"
)

const (
	apiport = 34001
)

func setup(t *testing.T) *sdk.Client {
	cf := mock.Cfg(t)
	cf.Port = apiport
	server := httptest.NewServer(svc.NewLogMux(cf))
	return sdk.TestClient(t, server.Client(), apiport)
}

func TestTestSetupWorks(t *testing.T) {
	setup(t)
}
