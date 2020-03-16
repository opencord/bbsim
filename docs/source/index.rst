.. BBSim documentation master file, created by
   sphinx-quickstart on Fri Oct 25 12:03:42 2019.
   You can adapt this file completely to your liking, but it should at least
   contain the root `toctree` directive.

Welcome to BBSim's documentation!
=================================

.. toctree::
   :maxdepth: 2
   :caption: Contents:

   operations.rst
   onu-state-machine.rst
   olt-state-machine.rst
   development-dependencies.rst
   bbr.rst
   bbsimctl.rst
   api.rst


Quickstart
----------

BBSim (a.k.a. BroadBand Simulator) is a tool designed to emulate an `Openolt
<https://github.com/opencord/openolt>`_ compatible device.

In order to use BBSim you need to have:

- a Kubernetes cluster
- helm
- a working installation of VOLTHA

We strongly recommend the utilization of `kind-voltha
<https://github.com/ciena/kind-voltha>`_ to setup such environment.

Installation
------------

Once VOLTHA is up and running, you can deploy BBSim with this command:

.. code:: bash

    helm install -n bbsim cord/bbsim

If you need to specify a custom image for BBSim you can:

.. code:: bash

    helm install -n bbsim cord/bbsim --set images.bbsim.repository=bbsim --set images.bbsim.tag=candidate --set images.bbsim.pullPolicy=Never

The BBSim installation can be customized to emulate multiple ONUs and multiple
PON Ports:

.. code:: bash

    helm install -n bbsim cord/bbsim --set onu=8 --set pon=2

BBSim can also be configured to automatically start Authentication or DHCP:

.. code:: bash

   helm install -n bbsim cord/bbsim --set auth=true --set dhcp=true

Once BBSim is installed you can verify that it's running with:

.. code:: bash

    kubectl logs -n voltha -f $(kubectl get pods -n voltha | grep bbsim | awk '{print $1}')

Provision a BBSim OLT in VOLTHA
-------------------------------

Create the device:

.. code:: bash

    voltctl device create -t openolt -H $(kubectl get -n voltha service/bbsim -o go-template='{{.spec.clusterIP}}'):50060

Enable the device:

.. code:: bash

    voltctl device enable $(voltctl device list --filter Type~openolt -q)

BBSim startup options
---------------------

``BBSim`` supports a series of options that can be set at startup, you can see
the list via ``./bbsim --help``

.. code:: bash

   $ ./bbsim --help
   Usage of ./bbsim:
     -auth
           Set this flag if you want authentication to start automatically
     -c_tag int
           C-Tag starting value, each ONU will get a sequential one (targeting 1024 ONUs per BBSim instance the range is big enough) (default 900)
     -cpuprofile string
           write cpu profile to file
     -delay int
           The delay between ONU DISCOVERY batches in milliseconds (1 ONU per each PON PORT at a time (default 200)
     -dhcp
           Set this flag if you want DHCP to start automatically
     -logCaller
           Whether to print the caller filename or not
     -logLevel string
           Set the log level (trace, debug, info, warn, error) (default "debug")
     -nni int
           Number of NNI ports per OLT device to be emulated (default 1)
     -olt_id int
           Number of OLT devices to be emulated
     -onu int
           Number of ONU devices per PON port to be emulated (default 1)
     -pon int
           Number of PON ports per OLT device to be emulated (default 1)
     -s_tag int
           S-Tag value (default 900)
     -enableEvents
           Set this flag for publishing BBSim events on configured kafkaAddress
     -kafkaAddress string
           IP:Port for kafka, used only when bbsimEvents flag is set (default ":9092")
     -ca string
           Set the mode for controlled activation of PON ports and ONUs
     -enableperf bool
           Setting this flag will cause BBSim to not store data like traffic schedulers, flows of ONUs etc

``BBSim`` also looks for a configuration file in ``configs/bbsim.yaml`` from
which it reads a number of default settings. The command line options listed
above override the corresponding configuration file settings. A sample
configuration file is given below:

.. literalinclude:: ../../configs/bbsim.yaml

Using the BBSim Sadis server in ONOS
------------------------------------

BBSim provides a simple server for testing with the ONOS Sadis app. The server
listens on port 50074 by default and provides the endpoints
``subscribers/<id>`` and ``bandwidthprofiles/<id>``.

To configure ONOS to use the BBSim ``Sadis`` server endpoints, the Sadis app
must use be configured as follows (see ``examples/sadis-in-bbsim.json``):

.. literalinclude:: ../../examples/sadis-in-bbsim2.json

This base configuration may also be obtained directly from the BBSim Sadis
server:

.. code:: bash

   curl http://<BBSIM_IP>:50074/cfg -o examples/sadis.json

It can then be pushed to the Sadis app using the following command:

.. code:: bash

   curl -sSL --user karaf:karaf \
       -X POST \
       -H Content-Type:application/json \
       http://localhost:8181/onos/v1/network/configuration/apps/org.opencord.sadis \
       --data @examples/sadis-in-bbsim.json

You can verify the current Sadis configuration:

.. code:: bash

   curl --user karaf:karaf http://localhost:8181/onos/v1/network/configuration/apps/org.opencord.sadis

In ONOS subscriber information can be queried using ``sadis <id>``.

*Note that BBSim supports both sadis configuration versions, here is an example of the configuration needed to return
the old format:*

.. literalinclude:: ../../examples/sadis-in-bbsim.json

Controlled PON and ONU activation
---------------------------------

BBSim provides support for controlled PON and ONU activation. By default both PON ports and ONUs are automatically activated when OLT is enabled.
This can be controlled using ``-ca`` option.

``-ca`` can be set to one of below four modes:

- default: PON ports and ONUs are automatic activated (default behavior).

- only-onu: PON ports automatically enabled and ONUs dynamically activated
            On Enable OLT, IntfIndications for all PON ports are sent but ONU discovery indications are not sent.
            When PoweronONU request is received at BBSim API server then ONU discovery indication is sent for that ONU.
- only-pon: PON ports dynamically enabled and ONUs automatically activated
            On Enable OLT, neither IntfIndications for PON ports nor ONU discovery indications are sent.
            When EnablePonIf request is received at OpenOLT server, then that PON port is enabled and
            IntfIndication is sent for that PON and all ONUs under that ports are discovered automatically.
- both: Both PON ports and ONUs are dynamically activated
        On Enable OLT, neither IntfIndication for PON ports nor ONU discovery indications are sent.
        When EnablePonIf request is received on OpenOLT server then IntfIndication is sent for that PON but no ONU discovery indication will be sent.
        When PoweronONU request is received at BBSim API server then ONU discovery indication is sent for that ONU.


Publishing BBSim Events on kafka
--------------------------------

BBSim provides option for publishing events on kafka. To publish events on kafka, set BBSimEvents flag and configure kafkaAddress.
Once BBSim is started, it will publish events (as and when they happen) on topic ``BBSim-OLT-<id>-Events``.

Types of Events:
  - OLT-enable-received
  - OLT-disable-received
  - OLT-reboot-received
  - OLT-reenable-received
  - ONU-discovery-indication-sent
  - ONU-activate-indication-received
  - MIB-upload-received
  - MIB-upload-done
  - Flow-add-received
  - Flow-remove-received
  - ONU-authentication-done
  - ONU-DHCP-ACK-received

Sample output of kafkacat consumer for BBSim with OLT-ID 4:

.. code:: bash

      $ kafkacat -b localhost:9092 -t BBSim-OLT-4-Events -C
      {"EventType":"OLT-enable-received","OnuSerial":"","OltID":4,"IntfID":-1,"OnuID":-1,"EpochTime":1583152243144,"Timestamp":"2020-03-02 12:30:43.144449453"}
      {"EventType":"ONU-discovery-indication-sent","OnuSerial":"BBSM00040001","OltID":4,"IntfID":0,"OnuID":0,"EpochTime":1583152243227,"Timestamp":"2020-03-02 12:30:43.227183506"}
      {"EventType":"ONU-activate-indication-received","OnuSerial":"BBSM00040001","OltID":4,"IntfID":0,"OnuID":1,"EpochTime":1583152243248,"Timestamp":"2020-03-02 12:30:43.248225467"}
      {"EventType":"MIB-upload-received","OnuSerial":"BBSM00040001","OltID":4,"IntfID":0,"OnuID":1,"EpochTime":1583152243299,"Timestamp":"2020-03-02 12:30:43.299480183"}
