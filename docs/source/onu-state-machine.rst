.. _ONU State Machine:

ONU State Machine
=================

In ``BBSim`` the device state is created using a state machine
library: `fsm <https://github.com/looplab/fsm>`__.

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
    * - receive_eapol_flow
      - enabled, gem_port_added
      - eapol_flow_received
      -
    * - add_gem_port
      - enabled, eapol_flow_received
      - gem_port_added
      - We need to wait for both the flow and the gem port to come before moving to ``auth_started``
    * - start_auth
      - eapol_flow_received, gem_port_added, eap_start_sent, eap_response_identity_sent, eap_response_challenge_sent, eap_response_success_received, auth_failed, dhcp_ack_received, dhcp_failed
      - auth_started
      -
    * - eap_start_sent
      - auth_started
      - eap_start_sent
      -
    * - eap_response_identity_sent
      - eap_start_sent
      - eap_response_identity_sent
      -
    * - eap_response_challenge_sent
      - eap_response_identity_sent
      - eap_response_challenge_sent
      -
    * - eap_response_success_received
      - eap_response_challenge_sent
      - eap_response_success_received
      -
    * - auth_failed
      - auth_started, eap_start_sent, eap_response_identity_sent, eap_response_challenge_sent
      - auth_failed
      -
    * - start_dhcp
      - eap_response_success_received, dhcp_discovery_sent, dhcp_request_sent, dhcp_ack_received, dhcp_failed
      - dhcp_started
      -
    * - dhcp_discovery_sent
      - dhcp_started
      - dhcp_discovery_sent
      -
    * - dhcp_request_sent
      - dhcp_discovery_sent
      - dhcp_request_sent
      -
    * - dhcp_ack_received
      - dhcp_request_sent
      - dhcp_ack_received
      -
    * - dhcp_failed
      - dhcp_started, dhcp_discovery_sent, dhcp_request_sent
      - dhcp_failed
      -
    * - disable
      - enabled, eap_response_success_received, auth_failed, dhcp_ack_received, dhcp_failed
      - disabled
      -
    * - pon_disabled
      - enabled, gem_port_added, eapol_flow_received, eap_response_success_received, auth_failed, dhcp_ack_received, dhcp_failed
      - pon_disabled
      -

In addition some transition can be forced via the API,
check the previous table to verify when you can trigger those actions and
:ref:`BBSimCtl` for more informations about ``BBSimCtl``:

.. list-table:: API State Transitions
    :widths: 15 15 70
    :header-rows: 1

    * - BBSimCtl command
      - Transitions To
      - Notes
    * - shutdown
      - disable
      - Emulates a device shutdown. Sends a ``DyingGaspInd`` and then an ``OnuIndication{OperState: 'down'}``
    * - poweron
      - discover
      - Emulates a device power on. Sends a ``OnuDiscInd``
    * - auth_restart
      - start_auth
      - Forces the ONU to send a new ``EapStart`` packet.
    * - dhcp_restart
      - start_dhcp
      - Forces the ONU to send a new ``DHCPDiscovery`` packet.
    * - softReboot
      - disable
      - Emulates a device soft reboot. Sends ``LosIndication{status: 'on'}`` and then after reboot delay ``LosIndication{status: 'off'}``
    * - hardReboot
      - disable
      - Emulates a device hard reboot. Sends a ``DyingGaspInd`` and ``OnuIndication{OperState: 'down'}`` and raise necessary alarmIndications and after reboot delay sends ``OnuDiscInd`` and clear alarm indications.
    * - onuAlarms
      - TBD
      -

.. list-table::
    :widths: 15 15 70
    :header-rows: 1

    * - OpenOlt request
      - Transitions To
      - Notes
    * - ActivateOnu
      - enable
      - VOLTHA assigns ID to ONU and then BBSim sends a ``OnuIndication{OperState: 'up'}``
    * - DeleteOnu
      - initialized
      - Reset ONU-ID to zero and move ONU to initialized state so that its ready to poweron again when required

Below is a diagram of the state machine:

- In blue PON related states
- In green EAPOL related states
- In yellow DHCP related states
- In purple operator driven states

..
  TODO Evaluate http://blockdiag.com/en/seqdiag/examples.html

.. graphviz::

    digraph {
        rankdir=TB
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
            }
            gem_port_added

            {created, disabled} -> initialized -> discovered -> enabled
        }

        subgraph cluster_eapol {
            style=rounded
            style=dotted
            node [fillcolor="#e6ffc2"]

            eapol_flow_received
            auth_started
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

            eap_start_sent -> auth_started
            eap_response_identity_sent -> auth_started
            eap_response_challenge_sent -> auth_started

            eap_response_success_received -> auth_started
            auth_failed -> auth_started
        }

        subgraph cluster_dhcp {
            node [fillcolor="#fffacc"]
            style=rounded
            style=dotted

            dhcp_started
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

            dhcp_discovery_sent -> dhcp_started
            dhcp_request_sent -> dhcp_started
            dhcp_ack_received -> dhcp_started
            dhcp_failed -> dhcp_started
        }
        enabled -> gem_port_added -> eapol_flow_received -> auth_started
        enabled -> eapol_flow_received -> gem_port_added -> auth_started

        {dhcp_ack_received, dhcp_failed} -> auth_started

        eap_response_success_received -> dhcp_started

        eap_response_success_received -> disabled
        auth_failed -> disabled
        dhcp_ack_received -> disabled
        dhcp_failed -> disabled
        disabled -> enabled
    }
