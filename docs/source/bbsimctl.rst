.. _BBSimCtl:

BBSimCtl
========

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

``bbsimctl`` can be configured via a config file such as:

.. code:: bash

    $ cat ~/.bbsim/config
    apiVersion: v1
    server: 127.0.0.1:50070
    grpc:
      timeout: 10s

Example commands
----------------

.. code:: bash

    $ ./bbsimctl olt get
    ID    SERIALNUMBER    OPERSTATE    INTERNALSTATE
    0     BBSIM_OLT_0     up           enabled


    $ ./bbsimctl olt pons
    PON Ports for : BBSIM_OLT_0

    ID    OPERSTATE
    0     up
    1     up
    2     up
    3     up


    $ ./bbsimctl onu list
    PONPORTID    ID    SERIALNUMBER    STAG    CTAG    OPERSTATE    INTERNALSTATE
    0            1     BBSM00000001    900     900     up           eap_response_identity_sent
    0            2     BBSM00000002    900     901     up           eap_start_sent
    0            3     BBSM00000003    900     902     up           auth_failed
    0            4     BBSM00000004    900     903     up           auth_failed
    1            1     BBSM00000101    900     904     up           eap_response_success_received
    1            2     BBSM00000102    900     905     up           eap_response_success_received
    1            3     BBSM00000103    900     906     up           eap_response_challenge_sent
    1            4     BBSM00000104    900     907     up           auth_failed
    2            1     BBSM00000201    900     908     up           auth_failed
    2            2     BBSM00000202    900     909     up           eap_start_sent
    2            3     BBSM00000203    900     910     up           eap_response_identity_sent
    2            4     BBSM00000204    900     911     up           eap_start_sent
    3            1     BBSM00000301    900     912     up           eap_response_identity_sent
    3            2     BBSM00000302    900     913     up           auth_failed
    3            3     BBSM00000303    900     914     up           auth_failed
    3            4     BBSM00000304    900     915     up           auth_failed

Autocomplete
------------

``bbsimctl`` comes with autocomplete, just run:

.. code:: bash

    source <(bbsimctl completion bash)
