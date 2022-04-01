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
  "status": "OK_STATUS",
  "inventory": {
    "lastChange": "1970-01-01T00:00:00Z",
    "root": {
      "name": "BBSim-BBSIM_OLT_10",
      "children": [
        {
          "name": "sfp-plus-transceiver-cage-0",
          "class": "COMPONENT_TYPE_CONTAINER",
          "description": "cage",
          "parent": "BBSim-BBSIM_OLT_10",
          "children": [
            {
              "name": "sfp-plus-0",
              "class": "COMPONENT_TYPE_TRANSCEIVER",
              "description": "bbsim-transceiver",
              "parent": "sfp-plus-transceiver-cage-0",
              "children": [
                {
                  "name": "pon-0",
                  "class": "COMPONENT_TYPE_PORT",
                  "description": "bbsim-pon-port",
                  "parent": "sfp-plus-0",
                  "uuid": {
                    "uuid": "38c3cdcc-f428-370b-a08f-8c64a90552e3"
                  },
                  "portAttr": {
                    "protocol": "XGSPON"
                  }
                }
              ],
              "uuid": {
                "uuid": "2d0ab069-f461-34e5-a16a-5ebb1f26b1a7"
              },
              "transceiverAttr": {
                "formFactor": "SFP_PLUS",
                "transType": "XGSPON",
                "maxDistance": 10,
                "maxDistanceScale": "VALUE_SCALE_KILO",
                "rxWavelength": [
                  1270
                ],
                "txWavelength": [
                  1577
                ],
                "wavelengthScale": "VALUE_SCALE_NANO"
              }
            }
          ],
          "uuid": {
            "uuid": "2410ab0a-1f17-3326-9708-da5b62b48d85"
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
        },
        {
          "name": "Thermal/PSU/SystemPSU/0",
          "class": "COMPONENT_TYPE_POWER_SUPPLY",
          "description": "bbsim-psu",
          "serialNum": "bbsim-psu-serial-0",
          "mfgName": "bbsim-psu",
          "uuid": {
            "uuid": "ee5fa7a2-0707-3bfb-90dd-678b2d6711c6"
          },
          "state": {
            
          }
        }
      ],
      "serialNum": "BBSIM_OLT_10",
      "mfgName": "BBSim",
      "uri": {
        "uri": "XXX.YYY.ZZZ.AAA"
      },
      "uuid": {
        "uuid": "895f65ab-64a3-3957-a2db-568a168faa3a"
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
    "parent": "sfp-plus-transceiver-cage-pon-0",
    "uuid": {
      "uuid": "788ba741-507f-37b1-8e47-4a389460873b"
    },
    "transceiverAttr": {
      "formFactor": "SFP_PLUS",
      "transType": "XGSPON",
      "maxDistance": 10,
      "maxDistanceScale": "VALUE_SCALE_KILO",
      "rxWavelength": [
        1270
      ],
      "txWavelength": [
        1577
      ],
      "wavelengthScale": "VALUE_SCALE_NANO"
    }
  }
}
```
### StopManagingDevice API
``` sh
grpcurl -plaintext -d '{"name":"SomeOlt"}' 172.17.0.2:50075 dmi.NativeHWManagementService.StopManagingDevice{
  "status": "OK_STATUS"
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
### List NativeEventsManagementService APIs
``` sh
$ grpcurl -plaintext 172.17.0.2:50075 list dmi.NativeEventsManagementService
dmi.NativeEventsManagementService.ListEvents
dmi.NativeEventsManagementService.UpdateEventsConfiguration
```
### ListEvents API
``` sh
$ grpcurl -plaintext -d '{"uuid": {"uuid":"5295a1d5-a121-372e-b8dc-6f7eda83f0ba"}}' 172.17.0.2:50075 dmi.NativeEventsManagementService.ListEvents
{
  "status": "OK_STATUS",
  "events": {
    "items": [
      {
        "eventId": "EVENT_FAN_FAILURE",
        "isConfigured": true
      },
      {
        "eventId": "EVENT_CPU_TEMPERATURE_ABOVE_CRITICAL",
        "isConfigured": true,
        "thresholds": {
          "upper": {
            "high": {
              "intVal": "95"
            },
            "low": {
              "intVal": "90"
            }
          }
        }
      },
      {
        "eventId": "EVENT_PSU_FAILURE",
        "isConfigured": true
      }
    ]
  }
}

```

### UpdateEventsConfiguration API
For FAN Failure, event_id must be set to 300 in grpc curl since enum value is 300
Please refer enums values corresponding to events in
https://github.com/opencord/device-management-interface/blob/master/protos/dmi/hw_events_mgmt_service.proto
``` sh
$ grpcurl -plaintext -d '{"device_uuid": {"uuid":"5295a1d5-a121-372e-b8dc-6f7eda83f0ba"}, "changes": {"items":{"event_id":"300", "is_configured": "false"}}}' 172.17.0.2:50075 dmi.NativeEventsManagementService.UpdateEventsConfiguration
{
  "status": "OK_STATUS"
}
```

### SetLoggingEndpoint API
``` sh
$ grpcurl -plaintext -d '{"device_uuid": {"uuid": "5295a1d5-a121-372e-b8dc-6f7eda83f0ba"}, "logging_endpoint": "127.0.0.1:543", "logging_protocol":"syslog"}' 172.17.0.2:50075 dmi.NativeHWManagementService/SetLoggingEndpoint
{
  "status": "OK_STATUS"
}
```

### GetLoggingEndpoint API
``` sh
$ grpcurl -plaintext -d '{"uuid": {"uuid": "5295a1d5-a121-372e-b8dc-6f7eda83f0ba"}}' 172.17.0.2:50075 dmi.NativeHWManagementService/GetLoggingEndpoint
{
  "status": "OK_STATUS",
  "loggingEndpoint": "127.0.0.1:543",
  "loggingProtocol": "syslog"
}
```

### UpdateStartupConfiguration
``` sh
$ grpcurl -plaintext -d '{"device_uuid": {"uuid": "5295a1d5-a121-372e-b8dc-6f7eda83f0ba"}}' 172.17.0.2:50075 dmi.NativeSoftwareManagementService/UpdateStartupConfiguration
{
  "status": "OK_STATUS"
}
```

### GetStartupConfigurationInfo
``` sh
$ grpcurl -plaintext -d '{"device_uuid": {"uuid": "5295a1d5-a121-372e-b8dc-6f7eda83f0ba"}}' 172.17.0.2:50075 dmi.NativeSoftwareManagementService/GetStartupConfigurationInfo
{
  "status": "OK_STATUS",
  "version": "BBSIM-STARTUP-CONFIG-DUMMY-VERSION"
}
```

## Generate DMI Events
Access bbsimctl
+++++++++++++++

When running a test you can check the state of each ONU using ``BBSimCtl``.

The easiest way to use ``bbsimctl`` is to ``exec`` inside the ``bbsim`` container:

.. code:: bash

    kubectl -n voltha exec -it $(kubectl -n voltha get pods -l app=bbsim -o name) -- /bin/bash

In case you prefer to run ``bbsimctl`` on your machine,
it can be configured via a config file such as:

.. code:: bash

    $ cat ~/.bbsim/config
    apiVersion: v1
    server: 127.0.0.1:50070
    grpc:
      timeout: 10s

Note : For event names, refer EventIds enum in https://github.com/opencord/device-management-interface/blob/master/protos/dmi/hw_events_mgmt_service.proto .

$ bbsimctl dmi events raise <event_name>

###  FAN FAILURE EVENT

.. code:: bash
```
$ bbsimctl dmi events create EVENT_FAN_FAILURE
[Status: 0] DMI Event Indication Sent.
```

###  PSU FAILURE EVENT

.. code:: bash
```
$ bbsimctl dmi events create EVENT_PSU_FAILURE
[Status: 0] DMI Event Indication Sent.
```

###  HW DEVICE TEMPERATURE ABOVE CRITICAL EVENT

.. code:: bash
```
$ bbsimctl dmi events create EVENT_HW_DEVICE_TEMPERATURE_ABOVE_CRITICAL
[Status: 0] DMI Event Indication Sent.
```

## Plug transceivers in or out

Access ``bbsimctl``, as described in the previous section

###  List transceivers
.. code:: bash
```
$ bbsimctl dmi transceiver list
ID    UUID                                    NAME          TECHNOLOGY    PLUGGEDIN    PONIDS
0     2d0ab069-f461-34e5-a16a-5ebb1f26b1a7    sfp-plus-0    XGSPON        false        [0]
```

### Plug transceiver out
.. code:: bash
```
$ bbsimctl dmi transceiver plug_out 0
[Status: 0] Plugged out transceiver 0
```

### Plug transceiver in
.. code:: bash
```
$ bbsimctl dmi transceiver plug_in 0
[Status: 0] Plugged in transceiver 0
```