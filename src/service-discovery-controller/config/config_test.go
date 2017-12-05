package config_test

import (
	"encoding/json"
	"fmt"
	. "service-discovery-controller/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	Context("when created from valid JSON", func() {
		It("contains the values in the JSON", func() {
			configJSON := []byte(`{
				"address":"example.com",
				"port":"80053",
				"index":"62",
				"server_cert": "some_path_server_cert",
				"server_key": "some_path_server_key",
				"ca_cert": "some_path_ca_cert",
				"nats":[
					{
						"host": "a-nats-host",
						"port": 1,
						"user": "a-nats-user",
						"pass": "a-nats-pass"
					},
					{
						"host": "b-nats-host",
						"port": 2,
						"user": "b-nats-user",
						"pass": "b-nats-pass"
					}
				],
				"staleness_threshold_seconds": 5,
				"pruning_interval_seconds": 3,
				"metrics_emit_seconds": 6,
				"metron_port": 8080
			}`)

			parsedConfig, err := NewConfig(configJSON)
			Expect(err).ToNot(HaveOccurred())

			Expect(parsedConfig.Address).To(Equal("example.com"))
			Expect(parsedConfig.Port).To(Equal("80053"))
			Expect(parsedConfig.Index).To(Equal("62"))
			Expect(parsedConfig.ServerCert).To(Equal("some_path_server_cert"))
			Expect(parsedConfig.ServerKey).To(Equal("some_path_server_key"))
			Expect(parsedConfig.CACert).To(Equal("some_path_ca_cert"))
			Expect(parsedConfig.Index).To(Equal("62"))
			Expect(parsedConfig.NatsServers()).To(ContainElement("nats://a-nats-user:a-nats-pass@a-nats-host:1"))
			Expect(parsedConfig.NatsServers()).To(ContainElement("nats://b-nats-user:b-nats-pass@b-nats-host:2"))
			Expect(parsedConfig.StalenessThresholdSeconds).To(Equal(5))
			Expect(parsedConfig.PruningIntervalSeconds).To(Equal(3))
			Expect(parsedConfig.MetricsEmitSeconds).To(Equal(6))
		})
	})

	Context("when constructed with invalid JSON", func() {
		It("returns an error", func() {
			configJSON := []byte(`garbage`)
			_, err := NewConfig(configJSON)
			Expect(err).To(MatchError(ContainSubstring("unmarshal config")))
		})
	})

	var requiredFields map[string]interface{}
	BeforeEach(func() {
		requiredFields = map[string]interface{}{
			"metron_port":                 8080,
			"staleness_threshold_seconds": 5,
			"pruning_interval_seconds":    3,
			"metrics_emit_seconds":        678,
		}
	})

	DescribeTable("when config file field contains an invalid value",
		func(invalidField string, value interface{}, errorString string) {
			cfg := cloneMap(requiredFields)
			cfg[invalidField] = value

			cfgBytes, _ := json.Marshal(cfg)
			_, err := NewConfig(cfgBytes)

			Expect(err).To(MatchError(fmt.Sprintf("invalid config: %s", errorString)))
		},

		Entry("invalid metron_port", "metron_port", -2, "MetronPort: less than min"),
		Entry("invalid staleness_threshold_seconds", "staleness_threshold_seconds", -2, "StalenessThresholdSeconds: less than min"),
		Entry("invalid pruning_interval_seconds", "pruning_interval_seconds", -2, "PruningIntervalSeconds: less than min"),
		Entry("invalid metrics_emit_seconds", "metrics_emit_seconds", -2, "MetricsEmitSeconds: less than min"),
	)
})

func cloneMap(original map[string]interface{}) map[string]interface{} {
	new := map[string]interface{}{}
	for k, v := range original {
		new[k] = v
	}
	return new
}
