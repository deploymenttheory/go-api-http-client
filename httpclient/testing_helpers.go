package httpclient

import "go.uber.org/zap"

func standardSugaredLogger() *zap.SugaredLogger {
	logger, _ := zap.NewProduction()
	return logger.Sugar()
}
