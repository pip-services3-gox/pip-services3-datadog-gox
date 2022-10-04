package clients1

import (
	"time"

	cconf "github.com/pip-services3-gox/pip-services3-commons-gox/config"
	cconv "github.com/pip-services3-gox/pip-services3-commons-gox/convert"
	cdata "github.com/pip-services3-gox/pip-services3-commons-gox/data"
	cerr "github.com/pip-services3-gox/pip-services3-commons-gox/errors"
	cref "github.com/pip-services3-gox/pip-services3-commons-gox/refer"
	cauth "github.com/pip-services3-gox/pip-services3-components-gox/auth"
	rpcclient "github.com/pip-services3-gox/pip-services3-rpc-gox/clients"
	"golang.org/x/net/context"
)

type DataDogMetricsClient struct {
	*rpcclient.RestClient
	defaultConfig      *cconf.ConfigParams
	credentialResolver *cauth.CredentialResolver
}

func NewDataDogMetricsClient(config *cconf.ConfigParams) *DataDogMetricsClient {
	c := &DataDogMetricsClient{
		RestClient:         rpcclient.NewRestClient(),
		credentialResolver: cauth.NewEmptyCredentialResolver(),
	}

	c.defaultConfig = cconf.NewConfigParamsFromTuples(
		"connection.protocol", "https",
		"connection.host", "api.datadoghq.com",
		"connection.port", 443,
		"credential.internal_network", "true",
	)

	if config != nil {
		c.Configure(context.Background(), config)
	}
	c.BaseRoute = "api/v1"
	return c
}

func (c *DataDogMetricsClient) Configure(ctx context.Context, config *cconf.ConfigParams) {
	config = c.defaultConfig.Override(config)
	c.RestClient.Configure(ctx, config)
	c.credentialResolver.Configure(ctx, config)
}

func (c *DataDogMetricsClient) SetReferences(ctx context.Context, refs cref.IReferences) {
	c.RestClient.SetReferences(ctx, refs)
	c.credentialResolver.SetReferences(ctx, refs)
}

func (c *DataDogMetricsClient) Open(ctx context.Context, correlationId string) error {
	credential, err := c.credentialResolver.Lookup(ctx, correlationId)
	if err != nil {
		return err
	}

	if credential == nil || credential.AccessKey() == "" {
		err = cerr.NewConfigError(correlationId, "NO_ACCESS_KEY", "Missing access key in credentials")
		return err
	}
	if c.Headers.Value() == nil {
		c.Headers = cdata.NewEmptyStringValueMap()
	}
	c.Headers.SetAsObject("DD-API-KEY", credential.AccessKey())
	return c.RestClient.Open(context.Background(), correlationId)
}

func (c *DataDogMetricsClient) convertTags(tags map[string]string) string {
	if tags == nil {
		return ""
	}

	builder := ""

	for key, val := range tags {
		if builder != "" {
			builder += ","
		}
		builder += key + ":" + val
	}
	return builder
}

func (c *DataDogMetricsClient) convertPoints(points []DataDogMetricPoint) []interface{} {

	result := make([]interface{}, 0)
	for _, p := range points {
		point := make([]string, 2)
		var tm int64 = 0
		if p.Time.IsZero() {
			tm = time.Now().UTC().UnixNano() / (int64)(time.Second)
		} else {
			tm = p.Time.UnixNano() / (int64)(time.Second)
		}
		point[0] = cconv.StringConverter.ToString(tm)
		point[1] = cconv.StringConverter.ToString(p.Value)
		result = append(result, point)
	}

	return result
}

func (c *DataDogMetricsClient) convertMetric(metric DataDogMetric) map[string]interface{} {
	tags := metric.Tags

	if metric.Service != "" {
		if tags == nil {
			tags = make(map[string]string, 0)
		}
		tags["service"] = metric.Service
	}

	result := map[string]interface{}{
		"metric": metric.Metric,
		"points": c.convertPoints(metric.Points),
	}

	if metric.Type != "" {
		result["type"] = metric.Type
	} else {
		result["type"] = "gauge"
	}

	if tags != nil {
		result["tags"] = c.convertTags(tags)
	}
	if metric.Host != "" {
		result["host"] = metric.Host
	}
	if metric.Interval > 0 {
		result["interval"] = metric.Interval
	}

	return result
}

func (c *DataDogMetricsClient) convertMetrics(metrics []DataDogMetric) map[string]interface{} {

	series := make([]interface{}, 0)
	for _, metric := range metrics {
		series = append(series, c.convertMetric(metric))
	}
	return map[string]interface{}{
		"series": series,
	}
}

func (c *DataDogMetricsClient) SendMetrics(ctx context.Context, correlationId string, metrics []DataDogMetric) error {
	data := c.convertMetrics(metrics)

	// Commented instrumentation because otherwise it will never stop sending logs...
	// timing := c.Instrument(correlationId, "datadog.send_metrics");
	result, err := c.Call(ctx, "post", "series", correlationId, nil, data)
	//timing.EndTiming();
	_, err = c.InstrumentError(ctx, correlationId, "datadog.send_metrics", err, result)
	return err
}
