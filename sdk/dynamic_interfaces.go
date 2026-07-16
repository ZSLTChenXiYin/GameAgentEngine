package sdk

const (
	AuthorityInterfaceGameClientRequestData = "game_client_request_data"
	AuthorityQueryPlayerState               = "player_state"
	AuthorityQueryPlayerInventory           = "player_inventory"
	AuthorityQueryPlayerWallet              = "player_wallet"
	AuthorityQueryPlayerLocation            = "player_location"
	AuthorityQueryNPCLocation               = "npc_location"
	AuthorityQuerySceneState                = "scene_state"
	AuthorityQueryRoomState                 = "room_state"
	AuthorityQueryTaskState                 = "task_state"
	AuthorityQueryItemPresence              = "item_presence"
)

var defaultAuthorityQueryTypes = []string{
	AuthorityQueryPlayerState,
	AuthorityQueryPlayerInventory,
	AuthorityQueryPlayerWallet,
	AuthorityQueryPlayerLocation,
	AuthorityQueryNPCLocation,
	AuthorityQuerySceneState,
	AuthorityQueryRoomState,
	AuthorityQueryTaskState,
	AuthorityQueryItemPresence,
}

// NewInvokeContext returns a request-scoped invoke context ready for fluent edits.
func NewInvokeContext() *InvokeContext {
	return &InvokeContext{}
}

// EnsureContext returns the existing request context or initializes one.
func (r *InvokeRequest) EnsureContext() *InvokeContext {
	if r.Context == nil {
		r.Context = &InvokeContext{}
	}
	return r.Context
}

// AddDynamicInterfaces appends request-scoped external interfaces.
func (c *InvokeContext) AddDynamicInterfaces(interfaces ...DynamicInterface) *InvokeContext {
	if c == nil || len(interfaces) == 0 {
		return c
	}
	c.DynamicInterfaces = append(c.DynamicInterfaces, interfaces...)
	return c
}

// AddDynamicInterfaces appends request-scoped external interfaces to the request context.
func (r *InvokeRequest) AddDynamicInterfaces(interfaces ...DynamicInterface) *InvokeRequest {
	if r == nil || len(interfaces) == 0 {
		return r
	}
	r.EnsureContext().AddDynamicInterfaces(interfaces...)
	return r
}

// DynamicDataRequestOption customizes one request-scoped data request interface.
type DynamicDataRequestOption func(*DynamicInterface)

// DynamicActionOption customizes one request-scoped action interface.
type DynamicActionOption func(*DynamicInterface)

// NewDynamicDataRequest builds a request-scoped data request interface.
func NewDynamicDataRequest(id, externalInterface string, options ...DynamicDataRequestOption) DynamicInterface {
	di := DynamicInterface{
		ID:                id,
		Kind:              DynamicInterfaceDataRequest,
		ExternalInterface: externalInterface,
	}
	for _, option := range options {
		if option != nil {
			option(&di)
		}
	}
	return di
}

// NewDynamicAction builds a request-scoped action interface.
func NewDynamicAction(id, externalInterface string, options ...DynamicActionOption) DynamicInterface {
	di := DynamicInterface{
		ID:                id,
		Kind:              DynamicInterfaceAction,
		ExternalInterface: externalInterface,
	}
	for _, option := range options {
		if option != nil {
			option(&di)
		}
	}
	return di
}

// NewAuthorityDataRequest builds the standard game-client authority query interface.
func NewAuthorityDataRequest(id string, maxQueries int) DynamicInterface {
	options := []DynamicDataRequestOption{
		WithDescription("Query authoritative game-side state such as HP, inventory, money, quest, scene, occupancy, and immediate room state."),
		WithQueryTypes(defaultAuthorityQueryTypes...),
	}
	if maxQueries > 0 {
		options = append(options, WithMaxQueries(maxQueries))
	}
	return NewDynamicDataRequest(id, AuthorityInterfaceGameClientRequestData, options...)
}

// WithDescription sets the model-facing description.
func WithDescription(description string) DynamicDataRequestOption {
	return func(di *DynamicInterface) {
		if di != nil {
			di.Description = description
		}
	}
}

// WithActionDescription sets the model-facing description for an action interface.
func WithActionDescription(description string) DynamicActionOption {
	return func(di *DynamicInterface) {
		if di != nil {
			di.Description = description
		}
	}
}

// WithQueryTypes constrains the allowed query types.
func WithQueryTypes(queryTypes ...string) DynamicDataRequestOption {
	return func(di *DynamicInterface) {
		if di != nil {
			di.QueryTypes = append([]string(nil), queryTypes...)
		}
	}
}

// WithArgsSchema sets a structured argument schema.
func WithArgsSchema(schema map[string]any) DynamicDataRequestOption {
	return func(di *DynamicInterface) {
		if di != nil {
			di.ArgsSchema = schema
		}
	}
}

// WithActionArgsSchema sets a structured argument schema for an action interface.
func WithActionArgsSchema(schema map[string]any) DynamicActionOption {
	return func(di *DynamicInterface) {
		if di != nil {
			di.ArgsSchema = schema
		}
	}
}

// WithMaxQueries caps how many queries the model may issue through this interface.
func WithMaxQueries(maxQueries int) DynamicDataRequestOption {
	return func(di *DynamicInterface) {
		if di != nil {
			di.MaxQueries = maxQueries
		}
	}
}

// WithMaxCalls caps how many action calls the model may issue through this interface.
func WithMaxCalls(maxCalls int) DynamicActionOption {
	return func(di *DynamicInterface) {
		if di != nil {
			di.MaxCalls = maxCalls
		}
	}
}
