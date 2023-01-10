/*
 * Copyright 2018-2023 Open Networking Foundation (ONF) and the ONF Contributors

 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at

 * http://www.apache.org/licenses/LICENSE-2.0

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package dmiserver

import (
	"context"
	"math/rand"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"

	dmi "github.com/opencord/device-management-interface/go/dmi"
)

//MetricGenerationFunc to generate the metrics to the kafka bus
type MetricGenerationFunc func(*dmi.Component, *DmiAPIServer) *dmi.Metric

// MetricTriggerConfig is the configuration of a metric and the time at which it will be exported
type MetricTriggerConfig struct {
	cfg dmi.MetricConfig
	t   time.Time
}

//DmiMetricsGenerator has the attributes for generating metrics
type DmiMetricsGenerator struct {
	apiSrv            *DmiAPIServer
	configuredMetrics map[dmi.MetricNames]MetricTriggerConfig
	access            sync.Mutex
	mgCancelFunc      context.CancelFunc
}

var dmiMG DmiMetricsGenerator

//StartMetricGenerator starts the metric generator
func StartMetricGenerator(apiSrv *DmiAPIServer) {
	log.Debugf("StartMetricGenerator invoked")
	// Seed the rand for use later on
	rand.Seed(time.Now().UnixNano())

	dmiMG = DmiMetricsGenerator{
		apiSrv: apiSrv,
	}
	dmiMG.configuredMetrics = make(map[dmi.MetricNames]MetricTriggerConfig)

	// Add CPU usage as a default Metric reported every 60 secs
	cpuMetricConfig := dmi.MetricConfig{
		MetricId:     dmi.MetricNames_METRIC_CPU_USAGE_PERCENTAGE,
		IsConfigured: true,
		PollInterval: 60,
	}
	dmiMG.configuredMetrics[dmi.MetricNames_METRIC_CPU_USAGE_PERCENTAGE] = MetricTriggerConfig{
		cfg: cpuMetricConfig,
		t:   time.Unix(0, 0),
	}

	// Add FAN speed metric as a default Metric reported every 120 secs
	fanSpeedMetricConfig := dmi.MetricConfig{
		MetricId:     dmi.MetricNames_METRIC_FAN_SPEED,
		IsConfigured: true,
		PollInterval: 120,
	}
	dmiMG.configuredMetrics[dmi.MetricNames_METRIC_FAN_SPEED] = MetricTriggerConfig{
		cfg: fanSpeedMetricConfig,
		t:   time.Unix(0, 0),
	}

	// Add RAM usage percentage metric reported every 60 seconds
	ramUsageMetricConfig := dmi.MetricConfig{
		MetricId:     dmi.MetricNames_METRIC_RAM_USAGE_PERCENTAGE,
		IsConfigured: false,
		PollInterval: 60,
	}
	dmiMG.configuredMetrics[dmi.MetricNames_METRIC_RAM_USAGE_PERCENTAGE] = MetricTriggerConfig{
		cfg: ramUsageMetricConfig,
		t:   time.Unix(0, 0),
	}

	// Add DISK usage percentage metric reported every 60 seconds
	diskUsageMetricConfig := dmi.MetricConfig{
		MetricId:     dmi.MetricNames_METRIC_DISK_USAGE_PERCENTAGE,
		IsConfigured: false,
		PollInterval: 60,
	}
	dmiMG.configuredMetrics[dmi.MetricNames_METRIC_DISK_USAGE_PERCENTAGE] = MetricTriggerConfig{
		cfg: diskUsageMetricConfig,
		t:   time.Unix(0, 0),
	}
	// Add Inner Surrounding TEMP usage percentage metric reported every 120 seconds
	innerTempUsageMetricConfig := dmi.MetricConfig{
		MetricId:     dmi.MetricNames_METRIC_INNER_SURROUNDING_TEMP,
		IsConfigured: true,
		PollInterval: 120,
	}
	dmiMG.configuredMetrics[dmi.MetricNames_METRIC_INNER_SURROUNDING_TEMP] = MetricTriggerConfig{
		cfg: innerTempUsageMetricConfig,
		t:   time.Unix(0, 0),
	}

	StartGeneratingMetrics()
}

// StartGeneratingMetrics starts the goroutine which submits metrics to the metrics channel
func StartGeneratingMetrics() {
	if dmiMG.apiSrv == nil {
		// Metric Generator is not yet initialized/started.
		// Means that the device is not managed on the DMI interface
		return
	}

	// initialize a new context
	var mgCtx context.Context
	mgCtx, dmiMG.mgCancelFunc = context.WithCancel(context.Background())

	go generateMetrics(mgCtx)
}

func generateMetrics(ctx context.Context) {
loop:
	for {
		select {
		case <-ctx.Done():
			log.Infof("Stopping generation of metrics ")
			break loop
		default:
			c := make(map[dmi.MetricNames]MetricTriggerConfig)

			dmiMG.access.Lock()
			for k, v := range dmiMG.configuredMetrics {
				c[k] = v
			}
			dmiMG.access.Unlock()

			now := time.Now()
			// For all the supported metrics
			for k, v := range c {
				if dmiMG.apiSrv.root == nil || dmiMG.apiSrv.root.Children == nil {
					// inventory might not yet be created or somehow disappeared
					break
				}

				if k == dmi.MetricNames_METRIC_CPU_USAGE_PERCENTAGE && v.cfg.IsConfigured {
					if now.Before(v.t) {
						continue
					}
					updateConfiguredMetrics(now, dmi.MetricNames_METRIC_CPU_USAGE_PERCENTAGE, &v)

					// Get the CPUs
					for _, cpu := range findComponentsOfType(dmiMG.apiSrv.root.Children, dmi.ComponentType_COMPONENT_TYPE_CPU) {
						m := generateCPUUsageMetric(cpu, dmiMG.apiSrv)
						logger.Debugf("Got metric %v", m)
						sendOutMetric(m, dmiMG.apiSrv)
					}
				} else if k == dmi.MetricNames_METRIC_FAN_SPEED && v.cfg.IsConfigured {
					if now.Before(v.t) {
						continue
					}
					updateConfiguredMetrics(now, dmi.MetricNames_METRIC_FAN_SPEED, &v)

					// Get the FANs
					for _, fan := range findComponentsOfType(dmiMG.apiSrv.root.Children, dmi.ComponentType_COMPONENT_TYPE_FAN) {
						m := generateFanSpeedMetric(fan, dmiMG.apiSrv)
						logger.Debugf("Got metric %v", m)
						sendOutMetric(m, dmiMG.apiSrv)
					}
				} else if k == dmi.MetricNames_METRIC_RAM_USAGE_PERCENTAGE && v.cfg.IsConfigured {
					if now.Before(v.t) {
						continue
					}
					updateConfiguredMetrics(now, dmi.MetricNames_METRIC_RAM_USAGE_PERCENTAGE, &v)
					// Get the RAM
					for _, ram := range findComponentsOfType(dmiMG.apiSrv.root.Children, dmi.ComponentType_COMPONENT_TYPE_MEMORY) {
						m := generateRAMUsageMetric(ram, dmiMG.apiSrv)
						logger.Debugf("Got metric for ram usage percentage %v", m)
						sendOutMetric(m, dmiMG.apiSrv)
					}
				} else if k == dmi.MetricNames_METRIC_DISK_USAGE_PERCENTAGE && v.cfg.IsConfigured {
					if now.Before(v.t) {
						continue
					}
					updateConfiguredMetrics(now, dmi.MetricNames_METRIC_DISK_USAGE_PERCENTAGE, &v)
					// Get the DISK
					for _, disk := range findComponentsOfType(dmiMG.apiSrv.root.Children, dmi.ComponentType_COMPONENT_TYPE_STORAGE) {
						m := generateDiskUsageMetric(disk, dmiMG.apiSrv)
						logger.Debugf("Got metric for disk usage percentage %v", m)
						sendOutMetric(m, dmiMG.apiSrv)
					}
				} else if k == dmi.MetricNames_METRIC_INNER_SURROUNDING_TEMP && v.cfg.IsConfigured {
					if now.Before(v.t) {
						continue
					}
					updateConfiguredMetrics(now, dmi.MetricNames_METRIC_INNER_SURROUNDING_TEMP, &v)
					// Get the INNER  SURROUNDING TEMPERATURE
					for _, isTemp := range findComponentsOfType(dmiMG.apiSrv.root.Children, dmi.ComponentType_COMPONENT_TYPE_SENSOR) {
						m := generateInnerSurroundingTempMetric(isTemp, dmiMG.apiSrv)
						logger.Debugf("Got metric for inner surrounding temperature %v", m)
						sendOutMetric(m, dmiMG.apiSrv)
					}
				}
			}
			time.Sleep(1 * time.Second)
		}
	}
}

func sendOutMetric(metric interface{}, apiSrv *DmiAPIServer) {
	select {
	case apiSrv.metricChannel <- metric:
	default:
		logger.Debugf("Channel not ready dropping Metric")
	}
}

func updateConfiguredMetrics(curr time.Time, typ dmi.MetricNames, old *MetricTriggerConfig) {
	dmiMG.access.Lock()
	dmiMG.configuredMetrics[typ] = MetricTriggerConfig{
		cfg: old.cfg,
		t:   curr.Add(time.Second * time.Duration(old.cfg.PollInterval)),
	}
	dmiMG.access.Unlock()
}

func updateMetricIDAndMetaData(id dmi.MetricNames, c *dmi.Component, apiSrv *DmiAPIServer, m *dmi.Metric) *dmi.Metric {
	m.MetricId = id
	m.MetricMetadata = &dmi.MetricMetaData{
		DeviceUuid:    apiSrv.uuid,
		ComponentUuid: c.Uuid,
		ComponentName: c.Name,
	}
	return m
}

func generateCPUUsageMetric(cpu *dmi.Component, apiSrv *DmiAPIServer) *dmi.Metric {
	var met dmi.Metric
	met = *updateMetricIDAndMetaData(dmi.MetricNames_METRIC_CPU_USAGE_PERCENTAGE, cpu, apiSrv, &met)
	met.Value = &dmi.ComponentSensorData{
		Value:     generateRand(1, 20),
		Type:      dmi.DataValueType_VALUE_TYPE_OTHER,
		Scale:     dmi.ValueScale_VALUE_SCALE_UNITS,
		Timestamp: timestamppb.Now(),
	}
	return &met
}

func generateFanSpeedMetric(fan *dmi.Component, apiSrv *DmiAPIServer) *dmi.Metric {
	var met dmi.Metric
	met = *updateMetricIDAndMetaData(dmi.MetricNames_METRIC_FAN_SPEED, fan, apiSrv, &met)
	met.Value = &dmi.ComponentSensorData{
		Value:     generateRand(3000, 4000),
		Type:      dmi.DataValueType_VALUE_TYPE_RPM,
		Scale:     dmi.ValueScale_VALUE_SCALE_UNITS,
		Timestamp: timestamppb.Now(),
	}
	return &met
}

// return a random number RAND which is:  lValue < RAND < hValue
func generateRand(lValue, hValue int32) int32 {
	if lValue >= hValue {
		return 0
	}

	diff := hValue - lValue

	randVal := rand.Int31n(diff)

	return lValue + randVal
}

//UpdateMetricConfig Adds/Updates the passed metric configuration
func UpdateMetricConfig(newCfg *dmi.MetricConfig) {
	dmiMG.access.Lock()
	dmiMG.configuredMetrics[newCfg.GetMetricId()] = MetricTriggerConfig{
		cfg: *newCfg,
		t:   time.Unix(0, 0),
	}
	dmiMG.access.Unlock()
	logger.Infof("Metric updated %v", newCfg)
}

func generateRAMUsageMetric(ram *dmi.Component, apiSrv *DmiAPIServer) *dmi.Metric {
	var met dmi.Metric
	met = *updateMetricIDAndMetaData(dmi.MetricNames_METRIC_RAM_USAGE_PERCENTAGE, ram, apiSrv, &met)
	met.Value = &dmi.ComponentSensorData{
		Value:     generateRand(1, 8),
		Type:      dmi.DataValueType_VALUE_TYPE_OTHER,
		Scale:     dmi.ValueScale_VALUE_SCALE_GIGA,
		Timestamp: timestamppb.Now(),
	}
	return &met
}

func generateDiskUsageMetric(disk *dmi.Component, apiSrv *DmiAPIServer) *dmi.Metric {
	var met dmi.Metric
	met = *updateMetricIDAndMetaData(dmi.MetricNames_METRIC_DISK_USAGE_PERCENTAGE, disk, apiSrv, &met)
	met.Value = &dmi.ComponentSensorData{
		Value:     generateRand(50, 500),
		Type:      dmi.DataValueType_VALUE_TYPE_OTHER,
		Scale:     dmi.ValueScale_VALUE_SCALE_GIGA,
		Timestamp: timestamppb.Now(),
	}
	return &met
}

func generateInnerSurroundingTempMetric(istemp *dmi.Component, apiSrv *DmiAPIServer) *dmi.Metric {
	var met dmi.Metric
	met = *updateMetricIDAndMetaData(dmi.MetricNames_METRIC_INNER_SURROUNDING_TEMP, istemp, apiSrv, &met)
	met.Value = &dmi.ComponentSensorData{
		Value:     generateRand(30, 40),
		Type:      dmi.DataValueType_VALUE_TYPE_CELSIUS,
		Scale:     dmi.ValueScale_VALUE_SCALE_UNITS,
		Timestamp: timestamppb.Now(),
	}
	return &met
}

// get the metrics list
func getMetricsList() []*dmi.MetricConfig {
	components := make(map[dmi.MetricNames]MetricTriggerConfig)
	dmiMG.access.Lock()

	for key, value := range dmiMG.configuredMetrics {
		components[key] = value
	}

	dmiMG.access.Unlock()

	var toRet []*dmi.MetricConfig
	for _, v := range components {
		metricConfig := v.cfg
		toRet = append(toRet, &metricConfig)
	}
	logger.Debugf("Metrics list %+v", toRet)
	return toRet
}

func getMetric(comp *dmi.Component, metricID dmi.MetricNames) *dmi.Metric {
	switch metricID {
	case dmi.MetricNames_METRIC_FAN_SPEED:
		metric := generateFanSpeedMetric(comp, dmiMG.apiSrv)
		return metric

	case dmi.MetricNames_METRIC_CPU_USAGE_PERCENTAGE:
		metric := generateCPUUsageMetric(comp, dmiMG.apiSrv)
		return metric

	case dmi.MetricNames_METRIC_RAM_USAGE_PERCENTAGE:
		metric := generateRAMUsageMetric(comp, dmiMG.apiSrv)
		return metric

	case dmi.MetricNames_METRIC_DISK_USAGE_PERCENTAGE:
		metric := generateDiskUsageMetric(comp, dmiMG.apiSrv)
		return metric

	case dmi.MetricNames_METRIC_INNER_SURROUNDING_TEMP:
		metric := generateInnerSurroundingTempMetric(comp, dmiMG.apiSrv)
		return metric
	}
	return nil
}

// StopGeneratingMetrics stops the goroutine which submits metrics to the metrics channel
func StopGeneratingMetrics() {
	if dmiMG.mgCancelFunc != nil {
		dmiMG.mgCancelFunc()
	}
}

// StopMetricGenerator stops the generation of metrics and cleans up all local context
func StopMetricGenerator() {
	logger.Debugf("StopMetricGenerator invoked")

	StopGeneratingMetrics()

	dmiMG.access.Lock()
	// reset it to an empty map
	dmiMG.configuredMetrics = make(map[dmi.MetricNames]MetricTriggerConfig)
	dmiMG.access.Unlock()
}
