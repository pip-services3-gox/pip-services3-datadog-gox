package clients_test

import (
	"context"
	"os"
	"testing"
	"time"

	cconf "github.com/pip-services3-gox/pip-services3-commons-gox/config"
	rnd "github.com/pip-services3-gox/pip-services3-commons-gox/random"
	clients1 "github.com/pip-services3-gox/pip-services3-datadog-gox/clients"
	"github.com/stretchr/testify/assert"
)

func TestDataDogMetricClient(t *testing.T) {
	var client *clients1.DataDogMetricsClient
	ctx := context.Background()

	apiKey := os.Getenv("DATADOG_API_KEY")
	if apiKey == "" {
		apiKey = "3eb3355caf628d4689a72084425177ac"
	}

	client = clients1.NewDataDogMetricsClient(nil)

	config := cconf.NewConfigParamsFromTuples(
		"source", "test",
		"credential.access_key", apiKey,
	)
	client.Configure(ctx, config)

	err := client.Open(ctx, "")
	assert.Nil(t, err)

	defer client.Close(ctx, "")

	t.Run("Send Metrics", func(t *testing.T) {
		metrics := []clients1.DataDogMetric{
			{
				Metric:  "test.metric.1",
				Service: "TestService Golang",
				Host:    "TestHost",
				Type:    clients1.Gauge,
				Points: []clients1.DataDogMetricPoint{
					{
						Time:  time.Now().UTC(),
						Value: rnd.Double.Next(0, 100),
					},
				},
			},
			{
				Metric:   "test.metric.2",
				Service:  "TestService Golang",
				Host:     "TestHost",
				Type:     clients1.Rate,
				Interval: 100,
				Points: []clients1.DataDogMetricPoint{
					{
						Time:  time.Now().UTC(),
						Value: rnd.Double.Next(0, 100),
					},
				},
			},
			{
				Metric:   "test.metric.3",
				Service:  "TestService Golang",
				Host:     "TestHost",
				Type:     clients1.Count,
				Interval: 100,
				Points: []clients1.DataDogMetricPoint{
					{
						Time:  time.Now().UTC(),
						Value: rnd.Double.Next(0, 100),
					},
				},
			},
		}

		err := client.SendMetrics(ctx, "", metrics)
		assert.Nil(t, err)

	})

}
