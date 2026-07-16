package sdk

import "strings"

const (
	ActionIDDataRequest = "data_request"
	ActionModeAsync     = "async"
)

// HasPendingDataRequest reports whether the response is paused on an async
// authority data request callback.
func (r *InvokeResponse) HasPendingDataRequest() bool {
	if r == nil {
		return false
	}
	for _, call := range r.ActionCalls {
		if strings.EqualFold(strings.TrimSpace(call.ActionID), ActionIDDataRequest) && strings.EqualFold(strings.TrimSpace(call.Mode), ActionModeAsync) {
			return true
		}
	}
	return false
}
