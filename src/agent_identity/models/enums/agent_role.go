package enums

import (
	"fmt"
	"strings"
)

type AgentRole string
type AgentRoleList []AgentRole

const (
	AgentRoleClone       AgentRole = "clone"
	AgentRoleHuman       AgentRole = "human"
	AgentRoleExpert      AgentRole = "expert"
	AgentRoleFacilitator AgentRole = "facilitator"
)

func GetAllAgentRoles() AgentRoleList {
	return AgentRoleList{
		AgentRoleClone,
		AgentRoleHuman,
		AgentRoleExpert,
		AgentRoleFacilitator,
	}
}

func (k AgentRole) String() string {
	return string(k)
}

func (k AgentRoleList) String() string {
	out := make([]string, 0, len(k))
	for _, v := range k {
		out = append(out, string(v))
	}
	return strings.Join(out, ", ")
}

func (k *AgentRole) Set(value string) error {
	normalizedValue := strings.ToLower(value)
	for _, v := range GetAllAgentRoles() {
		if normalizedValue == v.String() {
			*k = v
			return nil
		}
	}
	return fmt.Errorf("invalid AgentRole: %s", value)
}

func (k AgentRoleList) Contains(key AgentRole) bool {
	for _, v := range k {
		if v == key {
			return true
		}
	}
	return false
}
