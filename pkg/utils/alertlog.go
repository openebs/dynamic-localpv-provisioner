package utils

import (
	"log"

	"go.uber.org/zap"
)

var (
	// Logger facilitates logging with alert format
	Logger = initLogger()
)

func initLogger() *zap.SugaredLogger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	//logger, err := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	return logger.Sugar()
}
