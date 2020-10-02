/*
 * Copyright 2018-present Open Networking Foundation

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

	dmi "github.com/opencord/device-management-interface/go/dmi"
)

//ListMetrics lists the supported metrics for the passed device.
func (dms *DmiAPIServer) ListMetrics(ctx context.Context, req *dmi.HardwareID) (*dmi.ListMetricsResponse, error) {
	logger.Debugf("ListMetrics invoked with request %+v", req)
	//return empty list of metrics for now
	metrics := []*dmi.MetricConfig{{}}
	return &dmi.ListMetricsResponse{
		Status: dmi.Status_OK,
		Reason: 0,
		Metrics: &dmi.MetricsConfig{
			Metrics: metrics,
		},
	}, nil
}

//UpdateMetricsConfiguration updates the configuration of the list of metrics in the request
func (dms *DmiAPIServer) UpdateMetricsConfiguration(ctx context.Context, req *dmi.MetricsConfigurationRequest) (*dmi.MetricsConfigurationResponse, error) {
	logger.Debugf("UpdateMetricConfiguration invoked with request %+v", req)
	return &dmi.MetricsConfigurationResponse{
		Status: dmi.Status_OK,
	}, nil
}

//GetMetric gets the instantenous value of a metric
func (dms *DmiAPIServer) GetMetric(ctx context.Context, req *dmi.GetMetricRequest) (*dmi.GetMetricResponse, error) {
	logger.Debugf("GetMetric invoked with request %+v", req)
	return &dmi.GetMetricResponse{
		Status: dmi.Status_OK,
		Reason: dmi.Reason_UNDEFINED_REASON,
		Metric: &dmi.Metric{},
	}, nil

}
