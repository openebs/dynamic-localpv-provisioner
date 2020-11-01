/*
Copyright 2017 The OpenEBS Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package alertlog

import (
	"go.uber.org/zap"
	"log"
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

//sugar.Infow("failed to fetch URL",
//// Structured context as loosely typed key-value pairs.
//"url", url,
//"attempt", 3,
//"backoff", time.Second,
//)
//sugar.Infof("Failed to fetch URL: %s", url)
