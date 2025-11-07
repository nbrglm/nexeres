package models

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type OrgSettings struct {
	MFA OrgMFASettings `json:"mfa"`
}

type OrgMFASettings struct {
	Required bool `json:"required"`
	// role name to role-id mapping
	Roles map[string]uuid.UUID `json:"roles"`
}

func (s *OrgSettings) Scan(src any) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, s)
	case string:
		return json.Unmarshal([]byte(v), s)
	case nil:
		// No settings provided, use default
		*s = OrgSettings{}
		return nil
	default:
		return fmt.Errorf("cannot scan %T into OrgSettings", src)
	}
}
