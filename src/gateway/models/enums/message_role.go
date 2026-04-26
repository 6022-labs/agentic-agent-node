package enums

import "strings"

type MessageRole string
type MessageRoleList []MessageRole

const (
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
	MessageRoleSystem    MessageRole = "system"
)

func GetAllMessageRoles() MessageRoleList {
	return MessageRoleList{
		MessageRoleUser,
		MessageRoleAssistant,
		MessageRoleSystem,
	}
}

func (k MessageRoleList) String() string {
	var values []string
	for _, v := range k {
		values = append(values, string(v))
	}
	return strings.Join(values, ", ")
}

func (list MessageRoleList) Contains(value MessageRole) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}
