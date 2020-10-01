.. _BBSimAPI:

REST APIs
=========

BBSimAPI
--------

BBSim exposes gRPC and REST APIs for external control of the simulated OLT. The
API is defined using the protocol buffer specification in
`api/bbsim/bbsim.proto`.

By default, the gRPC server is started on port 50070 and the REST server is
started on port 50071. The following endpoints are currently defined:

.. openapi:: ../swagger/bbsim/bbsim.swagger.json

Legacy BBSimAPI
---------------

Additionally, a legacy API is available, defined in `api/legacy/bbsim.proto`.
This API is deprecated and will be removed once BBSim reaches feature parity
with the legacy version.

By default, the legacy gRPC server is started on port 50072 and the
corresponding REST server is started on port 50073. The following endpoints are
currently defined:

.. openapi:: ../swagger/legacy/bbsim.swagger.json
