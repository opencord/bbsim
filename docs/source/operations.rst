.. _Operations:

Operate BBSim
=============

If you are testing basic functionality using BBSim no operator intervention is required.

When you ``enable`` the device in VOLTHA the simulator will:

- activate all the configured ONUs
- wait for the EAPOL flow for each Service that requires it and trigger the authentication state machine as soon as
  it's received
- wait for the DHCP flow for each Service that requires it and trigger the DHCP state machine as soon as it's received

BBSimCtl
--------

BBSim comes with a gRPC interface to control the internal state. This
interface can be queried using `bbsimctl` (the tool can be build with
`make build` and it's available inside the `bbsim` container):

.. code:: bash

    $ ./bbsimctl --help
    Usage:
      bbsimctl [OPTIONS] <command>

    Global Options:
      -c, --config=FILE           Location of client config file [$BBSIMCTL_CONFIG]
      -s, --server=SERVER:PORT    IP/Host and port of XOS
      -d, --debug                 Enable debug mode

    Help Options:
      -h, --help                  Show this help message

    Available commands:
      completion  generate shell compleition
      config      generate bbsimctl configuration
      log         set bbsim log level
      olt         OLT Commands
      onu         ONU Commands
      pon         PON Commands
      service     Service Commands

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

Check the ONU Status
++++++++++++++++++++

.. code:: bash

    $ bbsimctl onu list
    PONPORTID    ID    PORTNO    SERIALNUMBER    OPERSTATE    INTERNALSTATE
    0            1     0         BBSM00000001    up           enabled

Check the Service Status
++++++++++++++++++++++++

.. code:: bash

    $ bbsimctl onu services BBSM00000001
    ONUSN           INTERNALSTATE    NAME    HWADDRESS            STAG    CTAG    NEEDSEAPOL    NEEDSDHCP    NEEDSIGMP    GEMPORT    EAPOLSTATE    DHCPSTATE            IGMPSTATE
    BBSM00000001    initialized      hsia    2e:60:00:00:01:00    900     900     false         false        false        1056       created       created              created
    BBSM00000001    initialized      voip    2e:60:00:00:01:01    333     444     false         true         false        1104       created       dhcp_ack_received    created
    BBSM00000001    initialized      vod     2e:60:00:00:01:02    555     55      false         true         true         1084       created       dhcp_ack_received    igmp_join_started
    BBSM00000001    initialized      MC      2e:60:00:00:01:03    550     55      false         false        false        0          created       created              created

Advanced operations
+++++++++++++++++++

In certain cases you may want to execute operations on the BBSim ONUs.

Here are the one currently supported, for more usage information use the following commands:

.. code:: bash

    $ bbsimctl onu --help
    Usage:
      bbsimctl [OPTIONS] onu <command>

    Commands to query and manipulate ONU devices

    Global Options:
      -c, --config=FILE           Location of client config file [$BBSIMCTL_CONFIG]
      -s, --server=SERVER:PORT    IP/Host and port of XOS
      -d, --debug                 Enable debug mode

    Help Options:
      -h, --help                  Show this help message

    Available commands:
      alarms
      auth_restart
      dhcp_restart
      flows
      get
      igmp
      invalidate_mds
      list
      poweron
      services
      shutdown
      traffic_schedulers

Enable ONUs
+++++++++++

``BBSimCtl`` gives you the ability to control the device lifecycle,
for example you can turn ONUs on and off:

.. code:: bash

    $ bbsimctl onu shutdown BBSM00000001
    [Status: 0] ONU BBSM00000001 successfully shut down.

    $ bbsimctl onu poweron BBSM00000001
    [Status: 0] ONU BBSM00000001 successfully powered on.

Autocomplete
++++++++++++

``bbsimctl`` comes with autocomplete, just run:

.. code:: bash

    source <(bbsimctl completion bash)

Other APIS
----------

.. toctree::
    :maxdepth: 1

    api.rst
