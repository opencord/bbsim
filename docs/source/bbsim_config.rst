BBSim Configuration
===================

BBSim startup options
---------------------

``BBSim`` supports a series of options that can be set at startup, you can see
the list via ``./bbsim --help``

.. code:: bash

   $ ./bbsim --help
   Usage of ./bbsim:
     -api_address string
           IP address:port (default ":50070")
     -authRetry bool
           Set this flag if BBSim should retry EAPOL (Authentication) upon failure until success (default false)
     -bp_format string
           Bandwidth profile format, 'mef' or 'ietf' (default "mef")
     -ca string
           Set the mode for controlled activation of PON ports and ONUs (default "default")
     -config string
           Configuration file path (default "configs/bbsim.yaml")
     -cpuprofile string
           write cpu profile to file (default "")
     -delay int
           The delay between ONU DISCOVERY batches in milliseconds (1 ONU per each PON PORT at a time (default 200)
     -dhcpRetry bool
           Set this flag if BBSim should retry DHCP upon failure until success (default false)
     -dmi_server_address string
           IP address:port (default ":50075")
     -enableEvents bool
           Enable sending BBSim events on configured kafka server (default false)
     -enableperf bool
           Setting this flag will cause BBSim to not store data like traffic schedulers, flows of ONUs etc.. (default false)
     -injectOmciUnknownAttributes bool
           Generate a MibDB packet with Unknown Attributes (default false)
     -injectOmciUnknownMe bool
           Generate a MibDB packet with Unknown Me (default false)
     -omccVersion int
           Set OMCC version to be returned in OMCI response of ME Onu2G (default 163 (0xA3))
     -kafkaAddress string
           IP:Port for kafka (default ":9092")
     -kafkaEventTopic string
           Ability to configure the topic on which BBSim publishes events on Kafka (default "")
     -logCaller bool
           Whether to print the caller filename or not (default false)
     -logLevel string
           Set the log level (trace, debug, info, warn, error) (default "debug")
     -nni int
           Number of NNI ports per OLT device to be emulated (default 1)
     -nni_speed uint
           Reported speed of the NNI ports in Mbps (default 1000)
     -oltRebootDelay int
           Time that BBSim should before restarting after a reboot (default 60)
     -olt_id int
           OLT device ID (default 0)
     -omci_response_rate int
           Amount of OMCI messages to respond to (default 10)
     -onu int
           Number of ONU devices per PON port to be emulated (default 1)
     -openolt_address string
           IP address:port (default ":50060")
     -pon int
           Number of PON ports per OLT device to be emulated (default 1)
     -pon_port_config_file string
        Pon Interfaces Configuration file path  (default "")
     -pots int
           Number of POTS UNI Ports per ONU device to be emulated (default 0)
     -rest_api_address string
           IP address:port (default ":50071")
     -services string
           Service Configuration file path (default "configs/att-services.yaml")
     -uni int
           Number of UNI Ports per ONU device to be emulated (default 4)




``BBSim`` also looks for a configuration file in ``configs/bbsim.yaml`` from
which it reads a number of default settings. The command line options listed
above override the corresponding configuration file settings. A sample
configuration file is given below:

.. literalinclude:: ../../configs/bbsim.yaml

.. _ConfiguringServices:

Configuring RG Services
-----------------------

BBSim supports different services in the RG.
Those services are described through a configuration file that is specified via the ``-services`` flag.

Below are examples of the tree commonly used configurations:

.. literalinclude:: ../../configs/att-services.yaml

.. literalinclude:: ../../configs/dt-services.yaml

.. literalinclude:: ../../configs/tt-services.yaml

Controlled PON and ONU activation
---------------------------------

BBSim provides support for controlled PON and ONU activation. This gives you the ability
to manually enable PON Ports and ONUs as needed.

By default both PON ports and ONUs are automatically activated when OLT is enabled.  This can
be controlled using ``-ca`` option.

``-ca`` can be set to one of below four modes:

- default: PON ports and ONUs are automatic activated (default behavior).

- only-onu: PON ports automatically enabled and ONUs activated on demand
            On Enable OLT, IntfIndications for all PON ports are sent but ONU discovery indications are not sent.
            When PoweronONU request is received at BBSim API server then ONU discovery indication is sent for that ONU.

- only-pon: PON ports enabled on demand and ONUs automatically activated
            On Enable OLT, neither IntfIndications for PON ports nor ONU discovery indications are sent.
            When EnablePonIf request is received at OpenOLT server, then that PON port is enabled and
            IntfIndication is sent for that PON and all ONUs under that ports are discovered automatically.

- both: Both PON ports and ONUs are activated on demand
        On Enable OLT, neither IntfIndication for PON ports nor ONU discovery indications are sent.
        When EnablePonIf request is received on OpenOLT server then
        IntfIndication is sent for that PON but no ONU discovery indication
        will be sent.
        When PoweronONU request is received at BBSim API server then ONU discovery indication is sent for that ONU.

Pon Port and ONU on demand activation is managed via ``BBSimCtl``.
You can find more information in the :ref:`Operations` page.

Publishing BBSim Events on kafka
--------------------------------

BBSim provides option for publishing events on kafka. To publish events on
kafka, set BBSimEvents flag and configure kafkaAddress.

Once BBSim is started, it will publish events (as and when they happen) on
topic ``BBSim-OLT-<id>-Events``.

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
