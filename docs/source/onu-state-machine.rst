.. _ONU State Machine:

ONU State Machine
=================

In ``BBSim`` the device state is createdtained using a state machine
library: `fsm <https://github.com/looplab/fsm>`__.

Here is a list of possible state transitions in BBSim:

+--------------------------------+----------------------------------------------------------------------------------------+--------------------------------+-----------------------------------------------------------------------------------------------+
| Transition                     | Starting States                                                                        | End State                      | Notes                                                                                         |
+================================+========================================================================================+================================+===============================================================================================+
|                                |                                                                                        | created                        |                                                                                               |
+--------------------------------+----------------------------------------------------------------------------------------+--------------------------------+-----------------------------------------------------------------------------------------------+
| discover                       | created                                                                                | discovered                     |                                                                                               |
+--------------------------------+----------------------------------------------------------------------------------------+--------------------------------+-----------------------------------------------------------------------------------------------+
| enable                         | discovered, disabled                                                                   | enabled                        |                                                                                               |
+--------------------------------+----------------------------------------------------------------------------------------+--------------------------------+-----------------------------------------------------------------------------------------------+
| receive_eapol_flow             | enabled, gem_port_added                                                                | eapol_flow_received            |                                                                                               |
+--------------------------------+----------------------------------------------------------------------------------------+--------------------------------+-----------------------------------------------------------------------------------------------+
| add_gem_port                   | enabled, eapol_flow_received                                                           | gem_port_added                 | We need to wait for both the flow and the gem port to come before moving to ``auth_started``  |
+--------------------------------+----------------------------------------------------------------------------------------+--------------------------------+-----------------------------------------------------------------------------------------------+
| start_auth                     | eapol_flow_received, gem_port_added                                                    | auth_started                   |                                                                                               |
+--------------------------------+----------------------------------------------------------------------------------------+--------------------------------+-----------------------------------------------------------------------------------------------+
| eap_start_sent                 | auth_started                                                                           | eap_start_sent                 |                                                                                               |
+--------------------------------+----------------------------------------------------------------------------------------+--------------------------------+-----------------------------------------------------------------------------------------------+
| eap_response_identity_sent     | eap_start_sent                                                                         | eap_response_identity_sent     |                                                                                               |
+--------------------------------+----------------------------------------------------------------------------------------+--------------------------------+-----------------------------------------------------------------------------------------------+
| eap_response_challenge_sent    | eap_response_identity_sent                                                             | eap_response_challenge_sent    |                                                                                               |
+--------------------------------+----------------------------------------------------------------------------------------+--------------------------------+-----------------------------------------------------------------------------------------------+
| eap_response_success_received  | eap_response_challenge_sent                                                            | eap_response_success_received  |                                                                                               |
+--------------------------------+----------------------------------------------------------------------------------------+--------------------------------+-----------------------------------------------------------------------------------------------+
| auth_failed                    | auth_started, eap_start_sent, eap_response_identity_sent, eap_response_challenge_sent  | auth_failed                    |                                                                                               |
+--------------------------------+----------------------------------------------------------------------------------------+--------------------------------+-----------------------------------------------------------------------------------------------+
| start_dhcp                     | eap_response_success_received                                                          | dhcp_started                   |                                                                                               |
+--------------------------------+----------------------------------------------------------------------------------------+--------------------------------+-----------------------------------------------------------------------------------------------+
| dhcp_discovery_sent            | dhcp_started                                                                           | dhcp_discovery_sent            |                                                                                               |
+--------------------------------+----------------------------------------------------------------------------------------+--------------------------------+-----------------------------------------------------------------------------------------------+
| dhcp_request_sent              | dhcp_discovery_sent                                                                    | dhcp_request_sent              |                                                                                               |
+--------------------------------+----------------------------------------------------------------------------------------+--------------------------------+-----------------------------------------------------------------------------------------------+
| dhcp_ack_received              | dhcp_request_sent                                                                      | dhcp_ack_received              |                                                                                               |
+--------------------------------+----------------------------------------------------------------------------------------+--------------------------------+-----------------------------------------------------------------------------------------------+
| dhcp_failed                    | dhcp_started, dhcp_discovery_sent, dhcp_request_sent                                   | dhcp_failed                    |                                                                                               |
+--------------------------------+----------------------------------------------------------------------------------------+--------------------------------+-----------------------------------------------------------------------------------------------+

In addition some transition can be forced via the API:

+---------------------+----------------------------------------------------------------------------+-----------+---------------------------------------------------------------------------------------------------------+
| End StateTransition | Starting States                                                            | End State | Notes                                                                                                   |
+=====================+============================================================================+===========+=========================================================================================================+
| disable             | eap_response_success_received, auth_failed, dhcp_ack_received, dhcp_failed | disabled  | Emulates a devide mulfunction. Sends a ``DyingGaspInd`` and then an ``OnuIndication{OperState: 'down'}``|
+---------------------+----------------------------------------------------------------------------+-----------+---------------------------------------------------------------------------------------------------------+

Below is a diagram of the state machine:

- In blue PON related states
- In green EAPOL related states
- In yellow DHCP related states
- In purple operator driven states

..
  TODO Evaluate http://blockdiag.com/en/seqdiag/examples.html

.. graphviz::

    digraph {
        graph [pad="1,1" bgcolor="#cccccc"]
        node [style=filled]

        created [fillcolor="#bee7fa"]
        discovered [fillcolor="#bee7fa"]
        enabled [fillcolor="#bee7fa"]
        disabled [fillcolor="#f9d6ff"]
        gem_port_added [fillcolor="#bee7fa"]

        eapol_flow_received [fillcolor="#e6ffc2"]
        auth_started [fillcolor="#e6ffc2"]
        eap_start_sent [fillcolor="#e6ffc2"]
        eap_response_identity_sent [fillcolor="#e6ffc2"]
        eap_response_challenge_sent [fillcolor="#e6ffc2"]
        eap_response_success_received [fillcolor="#e6ffc2"]
        auth_failed [fillcolor="#e6ffc2"]

        dhcp_started [fillcolor="#fffacc"]
        dhcp_discovery_sent [fillcolor="#fffacc"]
        dhcp_request_sent [fillcolor="#fffacc"]
        dhcp_ack_received [fillcolor="#fffacc"]
        dhcp_failed [fillcolor="#fffacc"]

        created -> discovered -> enabled
        enabled -> gem_port_added -> eapol_flow_received -> auth_started
        enabled -> eapol_flow_received -> gem_port_added -> auth_started

        auth_started -> eap_start_sent -> eap_response_identity_sent -> eap_response_challenge_sent -> eap_response_success_received
        auth_started -> auth_failed
        eap_start_sent -> auth_failed
        eap_response_identity_sent -> auth_failed
        eap_response_challenge_sent -> auth_failed

        eap_response_success_received -> dhcp_started
        dhcp_started -> dhcp_discovery_sent -> dhcp_request_sent -> dhcp_ack_received
        dhcp_started -> dhcp_failed
        dhcp_discovery_sent -> dhcp_failed
        dhcp_request_sent -> dhcp_failed
        dhcp_ack_received dhcp_failed

        eap_response_success_received -> disabled
        auth_failed -> disabled
        dhcp_ack_received -> disabled
        dhcp_failed -> disabled
        disabled -> enabled
    }