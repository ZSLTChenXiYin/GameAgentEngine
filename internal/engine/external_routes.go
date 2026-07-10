package engine

import (
	"strings"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/config"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/external"
)

func externalInterfaceConfig(name string) (config.ExternalInterfaceConfig, bool) {
	if strings.TrimSpace(name) == "" {
		return config.ExternalInterfaceConfig{}, false
	}
	cfg, ok := config.Global.ExternalInterfaces[strings.TrimSpace(name)]
	return cfg, ok
}

func gameClientInterfaceName(dr *DataRequest) string {
	if dr != nil && strings.TrimSpace(dr.ExternalInterface) != "" {
		return strings.TrimSpace(dr.ExternalInterface)
	}
	return "game_client_request_data"
}

func resolveGameClientRoute(dr *DataRequest) external.Route {
	interfaceName := gameClientInterfaceName(dr)
	base := config.ExternalInterfaceConfig{Consumer: "game_client"}
	if cfg, ok := externalInterfaceConfig(interfaceName); ok {
		base = cfg
	}
	if dr == nil {
		return external.NormalizeRouteWithOptions(base.DeliveryMode, base.PrimaryTransport, base.FallbackTransport, base.Consumer, base.ResumePolicy, base.TimeoutMs)
	}
	deliveryMode := firstNonEmpty(dr.DeliveryMode, base.DeliveryMode)
	primaryTransport := firstNonEmpty(dr.PrimaryTransport, base.PrimaryTransport)
	consumer := firstNonEmpty(dr.Consumer, base.Consumer, "game_client")
	timeoutMs := dr.TimeoutMs
	if timeoutMs <= 0 {
		timeoutMs = base.TimeoutMs
	}
	return external.NormalizeRouteWithOptions(deliveryMode, primaryTransport, base.FallbackTransport, consumer, base.ResumePolicy, timeoutMs)
}

func asyncActionInterfaceName(actionID string, args map[string]any) string {
	if args != nil {
		if raw, ok := args["external_interface"].(string); ok && strings.TrimSpace(raw) != "" {
			return strings.TrimSpace(raw)
		}
	}
	if strings.TrimSpace(actionID) == "" {
		return "async_action"
	}
	return strings.TrimSpace(actionID)
}

func resolveAsyncActionRoute(actionID string, args map[string]any) external.Route {
	interfaceName := asyncActionInterfaceName(actionID, args)
	base := config.ExternalInterfaceConfig{Consumer: "bridge"}
	if cfg, ok := externalInterfaceConfig(interfaceName); ok {
		base = cfg
	}
	deliveryMode := firstNonEmpty(runtimeTaskDeliveryModeFromArgs(args), base.DeliveryMode)
	primaryTransport := firstNonEmpty(runtimeTaskTransportFromArgs(args), base.PrimaryTransport)
	consumer := firstNonEmpty(runtimeTaskConsumerFromArgs(args), base.Consumer, "bridge")
	timeoutMs := runtimeTaskTimeoutFromArgs(args)
	if timeoutMs <= 0 {
		timeoutMs = base.TimeoutMs
	}
	return external.NormalizeRouteWithOptions(deliveryMode, primaryTransport, base.FallbackTransport, consumer, base.ResumePolicy, timeoutMs)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
