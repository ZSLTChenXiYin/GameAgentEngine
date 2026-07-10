package external

import (
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"testing"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/config"
)

type mockRPCDispatchService struct{}

func (s *mockRPCDispatchService) Dispatch(args RPCDispatchEnvelope, reply *DispatchResult) error {
	reply.Transport = args.Request.PrimaryTransport
	reply.Status = 202
	reply.Body = args.Auth.Mode + ":" + args.Auth.Token + ":" + args.Request.TaskID
	reply.Metadata = map[string]any{
		"method": "rpc",
		"task":   args.Request.TaskID,
	}
	return nil
}

func startRPCDispatchServer(t *testing.T) (string, func()) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen rpc server: %v", err)
	}
	server := rpc.NewServer()
	if err := server.RegisterName("Runtime", &mockRPCDispatchService{}); err != nil {
		listener.Close()
		t.Fatalf("register rpc service: %v", err)
	}
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go server.ServeCodec(jsonrpc.NewServerCodec(conn))
		}
	}()
	return "tcp://" + listener.Addr().String(), func() {
		_ = listener.Close()
		<-stopped
	}
}

func TestRPCAdapterDispatchCallsJSONRPCMethod(t *testing.T) {
	baseURL, stop := startRPCDispatchServer(t)
	defer stop()

	adapter := &RPCAdapter{}
	result, err := adapter.Dispatch(t.Context(), config.ExternalIntegrationConfig{
		Type:    "rpc_adapter",
		BaseURL: baseURL,
		Path:    "Runtime.Dispatch",
		Auth: config.ExternalIntegrationAuthConfig{
			Mode:  "bearer",
			Token: "rpc-secret",
		},
	}, DispatchRequest{TaskID: "task-rpc-1", PrimaryTransport: "game_rpc"})
	if err != nil {
		t.Fatalf("rpc dispatch: %v", err)
	}
	if result == nil || result.Status != 202 {
		t.Fatalf("unexpected rpc dispatch result: %+v", result)
	}
	if result.Body != "bearer:rpc-secret:task-rpc-1" {
		t.Fatalf("unexpected rpc dispatch body: %q", result.Body)
	}
}

func TestDispatcherDispatchUsesConfiguredRPCIntegration(t *testing.T) {
	baseURL, stop := startRPCDispatchServer(t)
	defer stop()

	previous := config.Global.ExternalIntegrations
	config.Global.ExternalIntegrations = map[string]config.ExternalIntegrationConfig{
		"game_rpc": {
			Type:    "rpc_adapter",
			BaseURL: baseURL,
			Path:    "Runtime.Dispatch",
		},
	}
	defer func() { config.Global.ExternalIntegrations = previous }()

	dispatcher := NewDispatcher()
	result, err := dispatcher.Dispatch(t.Context(), Route{DeliveryMode: "push", PrimaryTransport: "game_rpc"}, DispatchRequest{TaskID: "task-1", PrimaryTransport: "game_rpc"})
	if err != nil {
		t.Fatalf("rpc dispatch via dispatcher: %v", err)
	}
	if result == nil || result.Transport != "game_rpc" {
		t.Fatalf("unexpected dispatch result: %+v", result)
	}
}
