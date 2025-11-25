package logger

func LogQueryError(query string, err error) {
	if Logger == nil {
		return
	}
	Logger.Errorw("SQL query error",
		"query", query,
		"error", err,
	)
}

func LogTransactionStart(operation string) {
	if Logger == nil {
		return
	}
	Logger.Infow("transaction started",
		"operation", operation,
	)
}

func LogTransactionCommit(operation string) {
	if Logger == nil {
		return
	}
	Logger.Infow("transaction committed",
		"operation", operation,
	)
}

func LogTransactionRollback(operation string, err error) {
	if Logger == nil {
		return
	}
	Logger.Warnw("transaction rolled back",
		"operation", operation,
		"error", err,
	)
}
