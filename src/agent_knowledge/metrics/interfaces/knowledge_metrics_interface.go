package interfaces

import "context"

// KnowledgeMetricsInterface records conversation-knowledge activity (conversations started, messages appended).
type KnowledgeMetricsInterface interface {
	RecordConversationStarted(ctx context.Context)
	RecordMessageAppended(ctx context.Context, role string, count int64)
}
