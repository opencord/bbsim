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

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	dmi "github.com/opencord/device-management-interface/go/dmi"
)

//ListMetrics lists the supported metrics for the passed device.
func (dms *DmiAPIServer) ListMetrics(ctx context.Context, req *dmi.HardwareID) (*dmi.ListMetricsResponse, error) {
	logger.Debugf("ListMetrics invoked with request %+v", req)
	metrics := getMetricsList()

	return &dmi.ListMetricsResponse{
		Status: dmi.Status_OK_STATUS,
		Reason: 0,
		Metrics: &dmi.MetricsConfig{
			Metrics: metrics,
		},
	}, nil
}

//UpdateMetricsConfiguration updates the configuration of the list of metrics in the request
func (dms *DmiAPIServer) UpdateMetricsConfiguration(ctx context.Context, req *dmi.MetricsConfigurationRequest) (*dmi.MetricsConfigurationResponse, error) {
	logger.Debugf("UpdateMetricConfiguration invoked with request %+v", req)

	if req == nil || req.Operation == nil {
		return &dmi.MetricsConfigurationResponse{
			Status: dmi.Status_ERROR_STATUS,
			//TODO reason must be INVALID_PARAMS, currently this is available in Device Management interface (DMI),
			// change below reason with type INVALID_PARAMS once DMI is updated
			Reason: dmi.MetricsConfigurationResponse_INVALID_METRIC,
		}, status.Errorf(codes.FailedPrecondition, "request is nil")
	}

	switch x := req.Operation.(type) {
	case *dmi.MetricsConfigurationRequest_Changes:
		for _, chMetric := range x.Changes.Metrics {
			UpdateMetricConfig(chMetric)
		}
	case *dmi.MetricsConfigurationRequest_ResetToDefault:
		logger.Debugf("To be implemented later")
	case nil:
		// The field is not set.
		logger.Debugf("Update request operation type is nil")
		return &dmi.MetricsConfigurationResponse{
			Status: dmi.Status_UNDEFINED_STATUS,
		}, nil
	}

	return &dmi.MetricsConfigurationResponse{
		Status: dmi.Status_OK_STATUS,
	}, nil
}

//GetMetric gets the instantenous value of a metric
func (dms *DmiAPIServer) GetMetric(ctx context.Context, req *dmi.GetMetricRequest) (*dmi.GetMetricResponse, error) {
	logger.Debugf("GetMetric invoked with request %+v", req)

	if req == nil || req.GetMetricId() < 0 {
		return &dmi.GetMetricResponse{
			Status: dmi.Status_ERROR_STATUS,
			//TODO reason must be INVALID_PARAMS, currently this is not available in Device Management interface (DMI),
			// change below reason with type INVALID_PARAMS once DMI is updated
			Reason: dmi.GetMetricResponse_INVALID_METRIC,
			Metric: &dmi.Metric{},
		}, status.Errorf(codes.FailedPrecondition, "request is nil")
	}

	if dms.root == nil {
		return &dmi.GetMetricResponse{
			Status: dmi.Status_ERROR_STATUS,
			Reason: dmi.GetMetricResponse_INTERNAL_ERROR,
			Metric: &dmi.Metric{},
		}, status.Errorf(codes.FailedPrecondition, "Device is not managed, please start managing device to get the metric")
	}
	comp := findComponent(dms.root.Children, req.MetaData.ComponentUuid.Uuid)
	metric := getMetric(comp, req.GetMetricId())
	return &dmi.GetMetricResponse{
		Status: dmi.Status_OK_STATUS,
		Reason: dmi.GetMetricResponse_UNDEFINED_REASON,
		Metric: metric,
	}, nil
}

// Initiates the server streaming of the metrics
func (dms *DmiAPIServer) StreamMetrics(req *empty.Empty, srv dmi.NativeMetricsManagementService_StreamMetricsServer) error {
	return status.Errorf(codes.Unimplemented, "rpc StreamMetrics not implemented")
}
