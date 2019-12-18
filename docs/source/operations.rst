.. _Operations:

BBSim Operations
================

If you are testing basic functionality using BBSim no operator intervention is required.

When you ``enable`` the device in VOLTHA the simulator will:

- activate all the configured ONUs
- wait for the EAPOL flow for each ONU and trigger the authentication state machine as soon as it's received
- wait for the DHCP flow for each ONU and trigger the DHCP state machine as soon as it's received

Access bbsimctl
---------------

When running a test you can check the state of each ONU using :ref:`BBSimCtl`.

The easiest way to use ``bbsimctl`` is to ``exec`` inside the ``bbsim`` container:

.. code:: bash

    kubectl exec -it -n voltha -f $(kubectl get pods -n voltha | grep bbsim | awk '{print $1}') bash

Check the ONU Status
--------------------

.. code:: bash

    $ bbsimctl onu list
    PONPORTID    ID    PORTNO    SERIALNUMBER    HWADDRESS            STAG    CTAG    OPERSTATE    INTERNALSTATE
    0            1     0         BBSM00000001    2e:60:70:13:00:01    900     900     up           dhcp_ack_received

Advanced operations
-------------------

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
      auth_restart
      dhcp_restart
      get
      list
      poweron
      shutdown

