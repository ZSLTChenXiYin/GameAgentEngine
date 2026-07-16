package sdk

import "testing"

func TestInvokeResponseHasPendingDataRequest(t *testing.T) {
	resp := &InvokeResponse{ActionCalls: []ActionCall{{ActionID: ActionIDDataRequest, Mode: ActionModeAsync}}}
	if !resp.HasPendingDataRequest() {
		t.Fatal("expected pending data request detection")
	}
	resp = &InvokeResponse{ActionCalls: []ActionCall{{ActionID: "spawn_item", Mode: ActionModeAsync}}}
	if resp.HasPendingDataRequest() {
		t.Fatal("did not expect non-data_request action to be treated as pending data request")
	}
}
