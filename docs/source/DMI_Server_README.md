# Device Management Server simulator
The device management server can be tested using grpc curl (https://github.com/fullstorydev/grpcurl)
Below is the output of a run for some of the implemented RPCs (172.17.0.2 is the IP address of the BBSim docker)
### List NativeHWManagementService APIs:
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
### StartManagingDevice API
```sh
$ grpcurl -plaintext -d '{"name": "SomeOlt","uri": {"uri":"XXX.YYY.ZZZ.AAA"}}' 172.17.0.2:50075 dmi.NativeHWManagementService/StartManagingDevice
{
  "status": "OK",
  "deviceUuid": {
    "uuid": "5295a1d5-a121-372e-b8dc-6f7eda83f0ba"
  }
}
```
### GetPhysicalInventory API
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
        },
        {
          "name": "Thermal/Fans/System Fan/1",
          "class": "COMPONENT_TYPE_FAN",
          "description": "bbsim-fan",
          "serialNum": "bbsim-fan-serial-1",
          "mfgName": "bbsim-fan",
          "uuid": {
            "uuid": "2f4b639c-80a9-340f-b8d8-4ad06580b3cf"
          },
          "state": {

          }
        },
        {
          "name": "Thermal/Fans/System Fan/2",
          "class": "COMPONENT_TYPE_FAN",
          "description": "bbsim-fan",
          "serialNum": "bbsim-fan-serial-2",
          "mfgName": "bbsim-fan",
          "uuid": {
            "uuid": "4fb1a981-5697-3813-955d-fbb2a2908b2f"
          },
          "state": {

          }
        },
        {
          "name": "Systems/1/Disk/0",
          "class": "COMPONENT_TYPE_STORAGE",
          "description": "bbsim-disk",
          "serialNum": "bbsim-disk-serial-0",
          "mfgName": "bbsim-disk",
          "uuid": {
            "uuid": "b1d63346-c0b9-3a29-a4e6-e047efff9ddf"
          },
          "state": {

          }
        },
        {
          "name": "Systems/1/Processors/0",
          "class": "COMPONENT_TYPE_CPU",
          "description": "bbsim-cpu",
          "serialNum": "bbsim-cpu-serial-0",
          "mfgName": "bbsim-cpu",
          "uuid": {
            "uuid": "0bcbe296-9855-3218-9479-9c501073773f"
          },
          "state": {

          }
        },
        {
          "name": "Systems/1/Memory/0",
          "class": "COMPONENT_TYPE_MEMORY",
          "description": "bbsim-ram",
          "serialNum": "bbsim-ram-serial-0",
          "mfgName": "bbsim-ram",
          "uuid": {
            "uuid": "78cd16fd-73d1-3aad-8f13-644ed3101c63"
          },
          "state": {

          }
        },
        {
          "name": "Systems/1/Sensor/0",
          "class": "COMPONENT_TYPE_SENSOR",
          "description": "bbsim-istemp",
          "serialNum": "bbsim-sensor-istemp-serial-0",
          "mfgName": "bbsim-istemp",
          "uuid": {
            "uuid": "b9ada337-63a1-3991-96f5-95416dec1bf0"
          },
          "state": {

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
### GetHWComponentInfo API
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
### List NativeMetricsManagementService APIs
``` sh
$ grpcurl -plaintext 172.17.0.2:50075 list dmi.NativeMetricsManagementService
dmi.NativeMetricsManagementService.GetMetric
dmi.NativeMetricsManagementService.ListMetrics
dmi.NativeMetricsManagementService.UpdateMetricsConfiguration
```
### SetMsgBusEndpoint API
```sh
$ grpcurl -plaintext -d '{"msgbus_endpoint":"10.0.2.15:9092"}' 172.17.0.2:50075 dmi.NativeHWManagementService.SetMsgBusEndpoint
{
  "status": "OK"
}
```
### GetMsgBusEndpoint API
```sh
$ grpcurl -plaintext 172.17.0.2:50075 dmi.NativeHWManagementService.GetMsgBusEndpoint
{
  "status": "OK",
  "msgbusEndpoint": "10.0.2.15:9092"
}
```
### GetManagedDevices API
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
### GetMetric API for FAN speed
For FAN speed, metric_id must be set to 1 in grpc curl since enum value is 1.
```sh
$ grpcurl -plaintext -d '{"meta_data": {"device_uuid": {"uuid": "5295a1d5-a121-372e-b8dc-6f7eda83f0ba"},"component_uuid":{"uuid":"788ba741-507f-37b1-8e47-4a389460873b"}}, "metric_id": "1"}' 172.17.0.2:50075 dmi.NativeMetricsManagementService.GetMetric
{
  "status": "OK",
  "metric": {
    "metricId": "METRIC_FAN_SPEED",
    "metricMetadata": {
      "deviceUuid": {
        "uuid": "5295a1d5-a121-372e-b8dc-6f7eda83f0ba"
      },
      "componentUuid": {
        "uuid": "788ba741-507f-37b1-8e47-4a389460873b"
      },
      "componentName": "sfp-0"
    },
    "value": {
      "value": 3245,
      "type": "SENSOR_VALUE_TYPE_RPM",
      "scale": "SENSOR_VALUE_SCALE_UNITS",
      "timestamp": "2020-11-20T13:10:32.843964713Z"
    }
  }
}
```
### UpdateMetricsConfiguration API
``` sh
$ grpcurl -plaintext -d '{"device_uuid": {"uuid": "5295a1d5-a121-372e-b8dc-6f7eda83f0ba"}, "changes":{"metrics":{"metric_id":"1", "is_configured": "true", "poll_interval":"20"}}}' 172.17.0.2:50075 dmi.NativeMetricsManagementService.UpdateMetricsConfiguration
{
  "status": "OK"
}
```
### ListMetrics API
``` sh
$ grpcurl -plaintext -d '{"uuid": {"uuid":"5295a1d5-a121-372e-b8dc-6f7eda83f0ba"}}' 172.17.0.2:50075 dmi.NativeMetricsManagementService.ListMetrics
{
  "status": "OK",
  "metrics": {
    "metrics": [
      {
        "metricId": "METRIC_RAM_USAGE_PERCENTAGE",
        "isConfigured": true,
        "pollInterval": 60
      },
      {
        "metricId": "METRIC_DISK_USAGE_PERCENTAGE",
        "isConfigured": true,
        "pollInterval": 60
      },
      {
        "metricId": "METRIC_INNER_SURROUNDING_TEMP",
        "isConfigured": true,
        "pollInterval": 10
      },
      {
        "metricId": "METRIC_CPU_USAGE_PERCENTAGE",
        "isConfigured": true,
        "pollInterval": 5
      },
      {
        "metricId": "METRIC_FAN_SPEED",
        "isConfigured": true,
        "pollInterval": 10
      }
    ]
  }
}
```
