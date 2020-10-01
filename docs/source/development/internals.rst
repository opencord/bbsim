.. _BBSim Internals:

BBSim Internals
===============

.. toctree::
   :maxdepth: 1
   :caption: Other Resources:

   development-dependencies.rst

BBSim heavily leverages state machines to control the device lifecycle
and channels to propagate and react to state changes.

The most common pattern throughout the code is that any operations,
for example a gRPC call to the ``ActivateOnu`` endpoint will result in:

1. A state change in the ONU device, that will
2. Send a message on the ONU Channel, that will
3. Trigger some operation (for example send Indications to the OLT)

.. _OLT State Machine:

OLT State Machine
-----------------

Here is a list of possible states for an OLT:

.. list-table:: OLT States
    :header-rows: 1

    * -
      - Initialized
      - Enabled
      - Disabled
      - Deleted
    * - Data model is created for OLT, NNIs, PONs and ONUs
      - Starts the listener on the NNI interface and the DHCP server,
        Starts the OLT gRPC server,
        Moves the ONUs to ``initialized`` state
      - Sends OLT, NNIs and PONs ``UP`` indications
        Transition the ONUs into ``discovered`` state
      - Transition the ONUs into ``disabled`` state
        Sends OLT, NNIs and PONs ``UP`` indications
      - Stops the OLT gRPC Server

Below is a diagram of the state machine allowed transitions:

.. graphviz::

    digraph {
        rankdir=LR
        newrank=true
        graph [pad="1,1" bgcolor="#cccccc"]
        node [style=filled, fillcolor="#bee7fa"]

        created -> initialized -> enabled -> disabled -> deleted
        disabled -> enabled
        deleted -> initialized
    }



.. _ONU State Machine:

ONU State Machine
-----------------

Here is a list of possible state transitions for an ONU in BBSim:

.. list-table:: ONU States
    :widths: 10 35 10 45
    :header-rows: 1

    * - Transition
      - Starting States
      - End State
      - Notes
    * -
      -
      - created
      -
    * - initialize
      - created, disabled, pon_disabled
      - initialized
      -
    * - discover
      - initialized
      - discovered
      -
    * - enable
      - discovered, disabled, pon_disabled
      - enabled
      -
    * - disable
      - enabled
      - disabled
      - This state signifies that the ONU has been disabled
    * - pon_disabled
      - enabled
      - pon_disabled
      - This state signifies that the parent PON Port has been disabled, the ONU state hasn't been affected.

Below is a diagram of the state machine:

- In blue PON related states
- In purple operator driven states

.. graphviz::

    digraph {
        rankdir=LR
        newrank=true
        graph [pad="1,1" bgcolor="#cccccc"]
        node [style=filled]

        subgraph {
            node [fillcolor="#bee7fa"]

            created [peripheries=2]
            initialized
            discovered
            {
                rank=same
                enabled
                disabled [fillcolor="#f9d6ff"]
                pon_disabled [fillcolor="#f9d6ff"]
            }

            {created, disabled} -> initialized -> discovered -> enabled
        }

        disabled -> enabled
        enabled -> pon_disabled
        pon_disabled -> {initialized, disabled, enabled}
    }

.. _Service State Machine:

Service State Machine
---------------------

..
    TODO add table

.. graphviz::

    digraph {
        rankdir=TB
        newrank=true
        graph [pad="1,1" bgcolor="#cccccc"]
        node [style=filled]

        subgraph cluster_lifecycle {
            node [fillcolor="#bee7fa"]
            style=dotted

            created [peripheries=2]
            initialized
            disabled

            created -> initialized -> disabled
            disabled -> initialized
        }

        subgraph cluster_eapol {
            style=rounded
            style=dotted
            node [fillcolor="#e6ffc2"]

            auth_started [peripheries=2]
            eap_start_sent
            eap_response_identity_sent
            eap_response_challenge_sent
            {
                rank=same
                eap_response_success_received
                auth_failed
            }

            auth_started -> eap_start_sent -> eap_response_identity_sent -> eap_response_challenge_sent -> eap_response_success_received
            auth_started -> auth_failed
            eap_start_sent -> auth_failed
            eap_response_identity_sent -> auth_failed
            eap_response_challenge_sent -> auth_failed

            auth_failed -> auth_started
        }

        subgraph cluster_dhcp {
            node [fillcolor="#fffacc"]
            style=rounded
            style=dotted

            dhcp_started [peripheries=2]
            dhcp_discovery_sent
            dhcp_request_sent
            {
                rank=same
                dhcp_ack_received
                dhcp_failed
            }

            dhcp_started -> dhcp_discovery_sent -> dhcp_request_sent -> dhcp_ack_received
            dhcp_started -> dhcp_failed
            dhcp_discovery_sent -> dhcp_failed
            dhcp_request_sent -> dhcp_failed
            dhcp_ack_received dhcp_failed

        }

        subgraph cluster_igmp {
            node [fillcolor="#ffaaff"]
            style=rounded
            style=dotted

            igmp_join_started [peripheries=2]
            igmp_join_started -> igmp_join_error -> igmp_join_started
            igmp_join_started -> igmp_left
        }
    }
