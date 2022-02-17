package endpoints

import (
	"encoding/json"
	"fmt"
)

var _ DeliveryRuleCondition = DeliveryRuleSocketAddrCondition{}

type DeliveryRuleSocketAddrCondition struct {
	Parameters SocketAddrMatchConditionParameters `json:"parameters"`

	// Fields inherited from DeliveryRuleCondition
}

var _ json.Marshaler = DeliveryRuleSocketAddrCondition{}

func (s DeliveryRuleSocketAddrCondition) MarshalJSON() ([]byte, error) {
	type wrapper DeliveryRuleSocketAddrCondition
	wrapped := wrapper(s)
	encoded, err := json.Marshal(wrapped)
	if err != nil {
		return nil, fmt.Errorf("marshaling DeliveryRuleSocketAddrCondition: %+v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		return nil, fmt.Errorf("unmarshaling DeliveryRuleSocketAddrCondition: %+v", err)
	}
	decoded["name"] = "SocketAddr"

	encoded, err = json.Marshal(decoded)
	if err != nil {
		return nil, fmt.Errorf("re-marshaling DeliveryRuleSocketAddrCondition: %+v", err)
	}

	return encoded, nil
}