package enums

import "strings"

type LLMConnectionPlatform string
type LLMConnectionPlatformList []LLMConnectionPlatform

const (
	LLMConnectionPlatformX             LLMConnectionPlatform = "x"
	LLMConnectionPlatformMistral       LLMConnectionPlatform = "mistral"
	LLMConnectionPlatformOpenAI        LLMConnectionPlatform = "open-ai"
	LLMConnectionPlatformLibertAI      LLMConnectionPlatform = "libert-ai"
	LLMConnectionPlatformOllama        LLMConnectionPlatform = "ollama"
	LLMConnectionPlatformDeepseek      LLMConnectionPlatform = "deepseek"
	LLMConnectionPlatformAnthropic     LLMConnectionPlatform = "anthropic"
	LLMConnectionPlatformAmazonBedrock LLMConnectionPlatform = "amazon-bedrock"
)

func GetAllLLMConnectionPlatforms() LLMConnectionPlatformList {
	return LLMConnectionPlatformList{
		LLMConnectionPlatformX,
		LLMConnectionPlatformMistral,
		LLMConnectionPlatformOpenAI,
		LLMConnectionPlatformLibertAI,
		LLMConnectionPlatformOllama,
		LLMConnectionPlatformDeepseek,
		LLMConnectionPlatformAnthropic,
		LLMConnectionPlatformAmazonBedrock,
	}
}

func GetAllLocalLLMConnectionPlatforms() LLMConnectionPlatformList {
	return LLMConnectionPlatformList{
		LLMConnectionPlatformOllama,
	}
}

func GetAllCloudLLMConnectionPlatforms() LLMConnectionPlatformList {
	return LLMConnectionPlatformList{
		LLMConnectionPlatformX,
		LLMConnectionPlatformMistral,
		LLMConnectionPlatformOpenAI,
		LLMConnectionPlatformLibertAI,
		LLMConnectionPlatformDeepseek,
		LLMConnectionPlatformAnthropic,
		LLMConnectionPlatformAmazonBedrock,
	}
}

func (p LLMConnectionPlatform) String() string {
	return string(p)
}

func (l LLMConnectionPlatformList) String() string {
	out := make([]string, 0, len(l))
	for _, v := range l {
		out = append(out, string(v))
	}
	return strings.Join(out, ", ")
}

func (l LLMConnectionPlatformList) Contains(p LLMConnectionPlatform) bool {
	for _, v := range l {
		if v == p {
			return true
		}
	}
	return false
}
