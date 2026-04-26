package enums

import "strings"

type ConversationPlatform string
type ConversationPlatformList []ConversationPlatform

const (
	ConversationPlatformNone  ConversationPlatform = "none"
	ConversationPlatformSlack ConversationPlatform = "slack"
)

func GetAllConversationPlatforms() ConversationPlatformList {
	return ConversationPlatformList{
		ConversationPlatformNone,
		ConversationPlatformSlack,
	}
}

func (k ConversationPlatformList) String() string {
	var values []string
	for _, v := range k {
		values = append(values, string(v))
	}
	return strings.Join(values, ", ")
}

func (list ConversationPlatformList) Contains(value ConversationPlatform) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}
