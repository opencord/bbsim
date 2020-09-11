.. BBSim documentation master file, created by
   sphinx-quickstart on Fri Oct 25 12:03:42 2019.
   You can adapt this file completely to your liking, but it should at least
   contain the root `toctree` directive.

BBSim, a Broadband Simulator
============================

.. toctree::
   :maxdepth: 1
   :hidden:
   :caption: Contents:

   bbsim_config.rst
   sadis_config.rst
   operations.rst
   development/internals.rst
   Performances <bbr.rst>


Quickstart
----------

BBSim (a.k.a. BroadBand Simulator) is a tool designed to emulate an `Openolt
<https://github.com/opencord/openolt>`_ compatible device.

BBSim emulates the OLT, PON Ports, ONUs, UNIs and RG.
For more informations on how configure different Services on the RG refer to :ref:`ConfiguringServices`

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
