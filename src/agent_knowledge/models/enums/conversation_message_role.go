package enums

import "strings"

type ConversationMessageRole string
type ConversationMessageRoleList []ConversationMessageRole

const (
	ConversationMessageRoleUser      ConversationMessageRole = "user"
	ConversationMessageRoleAssistant ConversationMessageRole = "assistant"
	ConversationMessageRoleSystem    ConversationMessageRole = "system"
	ConversationMessageRoleTool      ConversationMessageRole = "tool"
)

func GetAllConversationMessageRoles() ConversationMessageRoleList {
	return ConversationMessageRoleList{
		ConversationMessageRoleUser,
		ConversationMessageRoleAssistant,
		ConversationMessageRoleSystem,
		ConversationMessageRoleTool,
	}
}

func (r ConversationMessageRole) String() string {
	return string(r)
}

func (l ConversationMessageRoleList) String() string {
	out := make([]string, 0, len(l))
	for _, r := range l {
		out = append(out, string(r))
	}
	return strings.Join(out, ", ")
}

func (l ConversationMessageRoleList) Contains(r ConversationMessageRole) bool {
	for _, v := range l {
		if v == r {
			return true
		}
	}
	return false
}
