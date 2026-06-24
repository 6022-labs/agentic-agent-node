package enums

import "strings"

type LLMConnectionPlatform string
type LLMConnectionPlatformList []LLMConnectionPlatform

// Values are the OTel gen_ai.provider.name well-known names so they map 1:1 to traces/metrics.
const (
	LLMConnectionPlatformXAi        LLMConnectionPlatform = "x_ai"
	LLMConnectionPlatformMistralAi  LLMConnectionPlatform = "mistral_ai"
	LLMConnectionPlatformOpenAi     LLMConnectionPlatform = "openai"
	LLMConnectionPlatformLibertAi   LLMConnectionPlatform = "libert_ai"
	LLMConnectionPlatformOllama     LLMConnectionPlatform = "ollama"
	LLMConnectionPlatformDeepseek   LLMConnectionPlatform = "deepseek"
	LLMConnectionPlatformAnthropic  LLMConnectionPlatform = "anthropic"
	LLMConnectionPlatformAwsBedrock LLMConnectionPlatform = "aws.bedrock"
)

func GetAllLLMConnectionPlatforms() LLMConnectionPlatformList {
	return LLMConnectionPlatformList{
		LLMConnectionPlatformXAi,
		LLMConnectionPlatformMistralAi,
		LLMConnectionPlatformOpenAi,
		LLMConnectionPlatformLibertAi,
		LLMConnectionPlatformOllama,
		LLMConnectionPlatformDeepseek,
		LLMConnectionPlatformAnthropic,
		LLMConnectionPlatformAwsBedrock,
	}
}

func GetAllLocalLLMConnectionPlatforms() LLMConnectionPlatformList {
	return LLMConnectionPlatformList{
		LLMConnectionPlatformOllama,
	}
}

func GetAllCloudLLMConnectionPlatforms() LLMConnectionPlatformList {
	return LLMConnectionPlatformList{
		LLMConnectionPlatformXAi,
		LLMConnectionPlatformMistralAi,
		LLMConnectionPlatformOpenAi,
		LLMConnectionPlatformLibertAi,
		LLMConnectionPlatformDeepseek,
		LLMConnectionPlatformAnthropic,
		LLMConnectionPlatformAwsBedrock,
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
