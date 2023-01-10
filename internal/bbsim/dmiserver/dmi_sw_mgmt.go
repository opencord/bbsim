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

	dmi "github.com/opencord/device-management-interface/go/dmi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//GetSoftwareVersion gets the software version information of the Active and Standby images
func (dms *DmiAPIServer) GetSoftwareVersion(ctx context.Context, req *dmi.HardwareID) (*dmi.GetSoftwareVersionInformationResponse, error) {
	// TODO: Make this more interesting by taking values from BBSim if available
	logger.Debugf("GetSoftwareVersion invoked with for device %+v", req)
	return &dmi.GetSoftwareVersionInformationResponse{
		Status: dmi.Status_OK_STATUS,
		Reason: dmi.GetSoftwareVersionInformationResponse_UNDEFINED_REASON,
		Info: &dmi.SoftwareVersionInformation{
			ActiveVersions: []*dmi.ImageVersion{{
				ImageName: "BBSIM-DUMMY-IMAGE-1",
				Version:   "BBSIM-DUMMY-VERSION-1",
			}},
			StandbyVersions: []*dmi.ImageVersion{{
				ImageName: "BBSIM-DUMMY-IMAGE-2",
				Version:   "BBSIM-DUMMY-VERSION-2",
			}},
		},
	}, nil
}

//DownloadImage downloads and installs the image in the standby partition, returns the status/progress of the Install
func (dms *DmiAPIServer) DownloadImage(req *dmi.DownloadImageRequest, stream dmi.NativeSoftwareManagementService_DownloadImageServer) error {
	logger.Debugf("DownloadImage invoked with request %+v", req)
	err := stream.Send(&dmi.ImageStatus{
		Status: dmi.Status_OK_STATUS,
		State:  dmi.ImageStatus_COPYING_IMAGE,
	})
	if err != nil {
		logger.Errorf("Error sending image-status for DownloadImage, %v", err)
		return err
	}
	return nil

}

//ActivateImage Activates and runs the OLT with the image in the standby partition
func (dms *DmiAPIServer) ActivateImage(req *dmi.HardwareID, stream dmi.NativeSoftwareManagementService_ActivateImageServer) error {
	logger.Debugf("ActivateImage invoked with request %+v", req)
	err := stream.Send(&dmi.ImageStatus{
		Status: dmi.Status_OK_STATUS,
		State:  dmi.ImageStatus_ACTIVATION_COMPLETE,
	})

	if err != nil {
		logger.Errorf("Error sending image-status for ActivateImage, %v", err)
		return err
	}
	return nil

}

//RevertToStandbyImage marks the image in the Standby as Active and reboots the device, so that it boots from that image which was in the standby.
func (dms *DmiAPIServer) RevertToStandbyImage(req *dmi.HardwareID, stream dmi.NativeSoftwareManagementService_RevertToStandbyImageServer) error {
	logger.Debugf("RevertToStandbyImage invoked with request %+v", req)
	err := stream.Send(&dmi.ImageStatus{
		Status: dmi.Status_OK_STATUS,
		State:  dmi.ImageStatus_ACTIVATION_COMPLETE,
	})

	if err != nil {
		logger.Errorf("Error sending image-status for RevertToStandbyImage, %v", err)
		return err
	}
	return nil
}

// UpdateStartupConfiguration API can be used to let the devices pickup their properitary configuration which they need at startup.
func (dms *DmiAPIServer) UpdateStartupConfiguration(request *dmi.ConfigRequest, stream dmi.NativeSoftwareManagementService_UpdateStartupConfigurationServer) error {
	logger.Debugf("UpdateStartupConfiguration invoked with request %+v", request)

	if request == nil {
		return status.Errorf(codes.InvalidArgument, "ConfigRequest is nil")
	}

	if request.DeviceUuid == nil || request.DeviceUuid.Uuid != dms.uuid.Uuid {
		if err := stream.Send(&dmi.ConfigResponse{
			Status: dmi.Status_ERROR_STATUS,
			Reason: dmi.ConfigResponse_UNKNOWN_DEVICE,
		}); err != nil {
			return status.Errorf(codes.Internal, "error sending response to client")
		}
		return nil
	}

	if err := stream.Send(&dmi.ConfigResponse{
		Status: dmi.Status_OK_STATUS,
	}); err != nil {
		return status.Errorf(codes.Internal, "error sending response to client")
	}

	return nil
}

// GetStartupConfigurationInfo API is used to return the 'StartUp' config present on the device
func (dms *DmiAPIServer) GetStartupConfigurationInfo(ctx context.Context, request *dmi.StartupConfigInfoRequest) (*dmi.StartupConfigInfoResponse, error) {
	logger.Debugf("GetStartupConfigurationInfo invoked for device %s", request.DeviceUuid.String())

	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "ConfigRequest is nil")
	}

	if request.DeviceUuid == nil {
		return nil, status.Errorf(codes.InvalidArgument, "DeviceUuid is nil")
	}

	if request.DeviceUuid.Uuid != dms.uuid.Uuid {
		return &dmi.StartupConfigInfoResponse{
			Status: dmi.Status_ERROR_STATUS,
			Reason: dmi.StartupConfigInfoResponse_UNKNOWN_DEVICE,
		}, status.Errorf(codes.InvalidArgument, "device-uuid %s not found", request.DeviceUuid.Uuid)
	}

	return &dmi.StartupConfigInfoResponse{
		Status:  dmi.Status_OK_STATUS,
		Version: "BBSIM-STARTUP-CONFIG-DUMMY-VERSION",
	}, nil
}
