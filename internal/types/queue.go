package types

import (
	"fmt"
)

const (
	TaskDeleteTempFile = "delete-temp-file"
)

type QueueDeleteTempFilePayload struct {
	FileName string `json:"file_name"`
	FilePath string `json:"path"`
}

type QueuePriority string

const (
	QueuePriorityCritical QueuePriority = "critical"
	QueuePriorityDefault  QueuePriority = "default"
	QueuePriorityLow      QueuePriority = "low"
)

func GetQueueName(p QueuePriority, env string) string {
	if env == "production" {
		return fmt.Sprintf("kelarin-%s-%s", p, env)
	}

	return fmt.Sprintf("kelarin-%s", p)
}

func GetQueuePriorityNameMap(env string) map[string]int {
	return map[string]int{
		GetQueueName(QueuePriorityCritical, env): 6,
		GetQueueName(QueuePriorityDefault, env):  3,
		GetQueueName(QueuePriorityLow, env):      1,
	}
}
