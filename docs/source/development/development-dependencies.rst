.. _Development dependencies:

Development dependencies
========================

If you want to test local changes in the ``omci-sim`` library you can
rebuild the BBSim container using a local version of the library with this command:

.. code:: bash

   LOCAL_OMCI_SIM=<path-to-omci-sim-library> make docker-build
