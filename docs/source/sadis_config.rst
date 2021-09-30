SADIS Configuration
===================

BBSim generates a ``sadis`` configuration starting from the :ref:`Service <ConfiguringServices>`
configuration file.

For example if you consider this Service configuration:

.. literalinclude:: ../../configs/att-services.yaml

The corresponding ``sadis`` configuration that you'll need to use in ONOS is:

.. literalinclude:: ../../examples/sadis-att.json

As you can imagine this file will get pretty big pretty quickly as soon as you scale
the number of ONUs and/or Services in your setup. For that reason (and to avoid mistakes)
we strongly suggest to configure ONOS to fetch informations from BBSim.

Configuring ONOS with an external sadis source is already done if you are using the `voltha-helm-charts
README <https://github.com/opencord/voltha-helm-charts/blob/master/README.md>`_.

If you prefer you can manually configure ONOS as explained below.

Using the BBSim Sadis server in ONOS
------------------------------------

BBSim provides a simple server for testing with the ONOS Sadis app. The server
listens on port 50074 by default and provides the endpoints
``subscribers/<id>`` and ``bandwidthprofiles/<id>``.

To configure ONOS to use the BBSim ``Sadis`` server endpoints, the Sadis app
must use be configured as follows (see ``examples/sadis-in-bbsim.json``):

.. literalinclude:: ../../examples/sadis-in-bbsim.json

This base configuration may also be obtained directly from the BBSim Sadis
server:

.. code:: bash

   curl http://<BBSIM_IP>:50074/v2/cfg -o examples/sadis.json

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

In ONOS subscriber information can be queried using ``sadis <uni-id>``.
