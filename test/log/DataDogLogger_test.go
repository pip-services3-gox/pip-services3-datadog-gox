package log_test

import (
	"context"
	"os"
	"testing"

	cconf "github.com/pip-services3-gox/pip-services3-commons-gox/config"
	ddlog "github.com/pip-services3-gox/pip-services3-datadog-gox/log"
	ddfixture "github.com/pip-services3-gox/pip-services3-datadog-gox/test/fixtures"

	"github.com/stretchr/testify/assert"
)

func TestDataDogLogger(t *testing.T) {
	ctx := context.Background()

	var logger *ddlog.DataDogLogger
	var fixture *ddfixture.LoggerFixture

	apiKey := os.Getenv("DATADOG_API_KEY")
	if apiKey == "" {
		apiKey = "3eb3355caf628d4689a72084425177ac"
	}

	logger = ddlog.NewDataDogLogger()
	fixture = ddfixture.NewLoggerFixture(logger.CachedLogger)

	config := cconf.NewConfigParamsFromTuples(
		"source", "test",
		"credential.access_key", apiKey,
	)
	logger.Configure(ctx, config)

	err := logger.Open(ctx, "")
	assert.Nil(t, err)

	defer logger.Close(ctx, "")

	t.Run("Log Level", func(t *testing.T) {
		fixture.TestLogLevel(t)
	})

	t.Run("Simple Logging", func(t *testing.T) {
		fixture.TestSimpleLogging(t)
	})

	t.Run("Error Logging", func(t *testing.T) {
		fixture.TestErrorLogging(t)
	})

}
