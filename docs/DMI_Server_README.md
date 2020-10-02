# Device Management Server simulator
The device management server can be tested using grpc curl (https://github.com/fullstorydev/grpcurl)
Below is the output of a run for some of the implemented RPCs (172.17.0.2 is the IP address of the BBSim docker)
```sh
$ grpcurl -plaintext 172.17.0.2:50075 list dmi.NativeHWManagementService
dmi.NativeHWManagementService.GetHWComponentInfo
dmi.NativeHWManagementService.GetLoggingEndpoint
dmi.NativeHWManagementService.GetManagedDevices
dmi.NativeHWManagementService.GetMsgBusEndpoint
dmi.NativeHWManagementService.GetPhysicalInventory
dmi.NativeHWManagementService.SetHWComponentInfo
dmi.NativeHWManagementService.SetLoggingEndpoint
dmi.NativeHWManagementService.SetMsgBusEndpoint
dmi.NativeHWManagementService.StartManagingDevice
dmi.NativeHWManagementService.StopManagingDevice
```
```sh
$ grpcurl -plaintext -d '{"name": "SomeOlt","uri": {"uri":"XXX.YYY.ZZZ.AAA"}}' 172.17.0.2:50075 dmi.NativeHWManagementService/StartManagingDevice
{
  "status": "OK",
  "deviceUuid": {
    "uuid": "5295a1d5-a121-372e-b8dc-6f7eda83f0ba"
  }
}
```
```sh
$ grpcurl -plaintext -d '{"device_uuid": {"uuid": "5295a1d5-a121-372e-b8dc-6f7eda83f0ba"}}' 172.17.0.2:50075 dmi.NativeHWManagementService.GetPhysicalInventory
{
  "status": "OK",
  "inventory": {
    "lastChange": "1970-01-01T00:00:00Z",
    "root": {
      "name": "BBSim-BBSIM_OLT_0",
      "children": [
        {
          "name": "cage-0",
          "class": "COMPONENT_TYPE_CONTAINER",
          "description": "cage",
          "parent": "BBSim-BBSIM_OLT_0",
          "children": [
            {
              "name": "sfp-0",
              "class": "COMPONENT_TYPE_TRANSCEIVER",
              "description": "XGS-PON",
              "parent": "cage-0",
              "uuid": {
                "uuid": "788ba741-507f-37b1-8e47-4a389460873b"
              }
            }
          ],
          "uuid": {
            "uuid": "6bfab16e-8003-3e18-9143-8933df64aa52"
          }
        }
      ],
      "serialNum": "BBSIM_OLT_0",
      "mfgName": "BBSim",
      "uri": {
        "uri": "XXX.YYY.ZZZ.AAA"
      },
      "uuid": {
        "uuid": "5295a1d5-a121-372e-b8dc-6f7eda83f0ba"
      },
      "state": {
      }
    }
  }
}
```
```sh
$ grpcurl -plaintext -d '{"device_uuid": {"uuid": "5295a1d5-a121-372e-b8dc-6f7eda83f0ba"},"component_uuid":{"uuid":"788ba741-507f-37b1-8e47-4a389460873b"}}' 172.17.0.2:50075 dmi.NativeHWManagementService.GetHWComponentInfo
{
  "status": "OK",
  "component": {
    "name": "sfp-0",
    "class": "COMPONENT_TYPE_TRANSCEIVER",
    "description": "XGS-PON",
    "parent": "cage-0",
    "uuid": {
      "uuid": "788ba741-507f-37b1-8e47-4a389460873b"
    }
  }
}
```
```sh
$ grpcurl -plaintext -d '{}' 172.17.0.2:50075 dmi.NativeHWManagementService.GetMsgBusEndpoint
ERROR:
  Code: Unimplemented
  Message: rpc GetMsgBusEndpoint not implemented
```
```sh
$ grpcurl -plaintext -d '{}' 172.17.0.2:50075 dmi.NativeHWManagementService/GetManagedDevices
{
  "devices": [
    {
      "name": "BBSim-BBSIM_OLT_0",
      "uri": {
        "uri": "XXX.YYY.ZZZ.AAA"
      }
    }
  ]
}
```