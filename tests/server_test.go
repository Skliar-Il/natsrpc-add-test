package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/LeKovr/natsrpc"
	"github.com/LeKovr/natsrpc/echopb"
)

type echoHandler struct {
	echopb.UnimplementedEchoServiceServer
}

func (h *echoHandler) Echo(ctx context.Context, req *echopb.EchoRequest) (*echopb.EchoReply, error) {
	return &echopb.EchoReply{Message: "echo: " + req.Message}, nil
}

func TestEchoIntegration(t *testing.T) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "nats:2.9.4-alpine",
		ExposedPorts: []string{"4222/tcp"},
		WaitingFor:   wait.ForLog("Server is ready"),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err, "failed to start NATS container")
	defer container.Terminate(ctx)

	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "4222")
	require.NoError(t, err)
	url := fmt.Sprintf("nats://%s:%s", host, port.Port())

	nc, err := nats.Connect(url, nats.Timeout(5*time.Second))
	require.NoError(t, err, "failed to connect to NATS")
	defer nc.Close()

	server, err := natsrpc.NewServer(nc)
	require.NoError(t, err, "failed to create natsrpc.Server")
	defer server.Close(context.Background())

	svc, err := echopb.RegisterEchoServiceNRServer(server, &echoHandler{})
	require.NoError(t, err, "failed to register EchoService")
	defer svc.Close()

	client := natsrpc.NewClient(nc)
	echoClient := echopb.NewEchoServiceNRClient(client)

	callCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	resp, err := echoClient.Echo(callCtx, &echopb.EchoRequest{Message: "hello"})
	require.NoError(t, err, "Echo RPC failed")
	require.Equal(t, "echo: hello", resp.Message)
}
