package enums

import "strings"

type LLMConnectionSetupStatus string
type LLMConnectionSetupStatusList []LLMConnectionSetupStatus

const (
	LLMConnectionSetupStatusEstablished LLMConnectionSetupStatus = "established"
)

func GetAllLLMConnectionSetupStatuses() LLMConnectionSetupStatusList {
	return LLMConnectionSetupStatusList{
		LLMConnectionSetupStatusEstablished,
	}
}

func (s LLMConnectionSetupStatus) String() string {
	return string(s)
}

func (l LLMConnectionSetupStatusList) String() string {
	out := make([]string, 0, len(l))
	for _, v := range l {
		out = append(out, string(v))
	}
	return strings.Join(out, ", ")
}

func (l LLMConnectionSetupStatusList) Contains(s LLMConnectionSetupStatus) bool {
	for _, v := range l {
		if v == s {
			return true
		}
	}
	return false
}
