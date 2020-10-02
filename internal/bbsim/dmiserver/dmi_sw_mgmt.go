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

//GetSoftwareVersion gets the software version information of the Active and Standby images
func (dms *DmiAPIServer) GetSoftwareVersion(ctx context.Context, req *dmi.HardwareID) (*dmi.GetSoftwareVersionInformationResponse, error) {
	// TODO: Make this more interesting by taking values from BBSim if available
	logger.Debugf("GetSoftwareVersion invoked with for device %+v", req)
	return &dmi.GetSoftwareVersionInformationResponse{
		Status: dmi.Status_OK,
		Reason: dmi.Reason_UNDEFINED_REASON,
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
		Status: dmi.Status_OK,
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
		Status: dmi.Status_OK,
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
		Status: dmi.Status_OK,
		State:  dmi.ImageStatus_ACTIVATION_COMPLETE,
	})

	if err != nil {
		logger.Errorf("Error sending image-status for RevertToStandbyImage, %v", err)
		return err
	}
	return nil
}
