package logger

import (
	"time"
)

func LogBusinessTransactionStart(operation string, fields map[string]interface{}) {
	if Logger == nil {
		return
	}
	capacity := 2 + len(fields)*2
	if capacity < 4 {
		capacity = 4
	}
	args := make([]interface{}, 0, capacity)
	args = append(args, "operation", operation)
	for k, v := range fields {
		args = append(args, k, v)
	}
	Logger.Infow("business transaction started", args...)
}

func LogBusinessTransactionEnd(operation string, duration time.Duration, success bool, fields map[string]interface{}) {
	if Logger == nil {
		return
	}
	capacity := 6 + len(fields)*2
	if capacity < 8 {
		capacity = 8
	}
	args := make([]interface{}, 0, capacity)
	args = append(args, "operation", operation, "duration", duration, "success", success)
	for k, v := range fields {
		args = append(args, k, v)
	}
	if success {
		Logger.Infow("business transaction completed", args...)
	} else {
		Logger.Warnw("business transaction failed", args...)
	}
}

func LogBusinessRule(rule string, fields map[string]interface{}) {
	if Logger == nil {
		return
	}
	capacity := 2 + len(fields)*2
	if capacity < 4 {
		capacity = 4
	}
	args := make([]interface{}, 0, capacity)
	args = append(args, "rule", rule)
	for k, v := range fields {
		args = append(args, k, v)
	}
	Logger.Debugw("business rule applied", args...)
}

func LogCriticalEvent(event string, fields map[string]interface{}) {
	if Logger == nil {
		return
	}
	capacity := 2 + len(fields)*2
	if capacity < 4 {
		capacity = 4
	}
	args := make([]interface{}, 0, capacity)
	args = append(args, "event", event)
	for k, v := range fields {
		args = append(args, k, v)
	}
	Logger.Infow("critical business event", args...)
}
