.. _OLT State Machine:

OLT State Machine
=================

In ``BBSim`` the device state is created using a state machine
library: `fsm <https://github.com/looplab/fsm>`__.

Here is a list of possible states for an OLT in BBSim:

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
        rankdir=TB
        newrank=true
        graph [pad="1,1" bgcolor="#cccccc"]
        node [style=filled]

        created -> initialized -> enabled -> disabled -> deleted
        disabled -> enabled
        deleted -> initialized
    }

