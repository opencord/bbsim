# BroadBand Simular (BBSim) 

## Build

BBSim is managed via a `Makefile`, plese run the following command
to display all the available targets

```
make help
```

## Usage

### Deploy on Kubernets

Once VOLTHA is deployed you can deploy BBsim using the helm chart provided in the repo:

```
git clone https://github.com/opencord/helm-charts.git
helm install -n bbsim bbsim
```

### OLT Provisioning

Once `BBSim` is up and running you can provision the OLT in VOLTHA.

When you install the `BBSim` helm chart you'll notice that the last line of the output
prints the service name and port:

```bash
...
NOTES:
BBSim deployed with release name: bbsim

OLT ID: 0
# of NNI Ports: 1
# of PON Ports: 2
# of ONU Ports: 2
Total ONUs: (total: 4)

OLT is listening on: "voltha.svc.bbsim:50060"
```

#### VOLTHA 1.X

Connect to the voltha CLI and execute this commands:

```bash
preprovision_olt -t openolt -H voltha.svc.bbsim:50060
enable
```

#### VOLTHA 2.X

_This assumes `voltctl` is installed an configured_

```bash
voltctl device create -t openolt -H bbsim:50060
voltctl device enable $(voltctl device list --filter Type~openolt -q)
```

## Control API

BBSim comes with a gRPC interface to control the internal state.
This interface can be queried using `bbsimctl` (the tool can be build with `make build`
and it's available inside the `bbsim` container):

```bash
$ ./bbsimctl --help
Usage:
  bbsimctl [OPTIONS] <config | olt | onus>

Global Options:
  -c, --config=FILE           Location of client config file [$BBSIMCTL_CONFIG]
  -s, --server=SERVER:PORT    IP/Host and port of XOS
  -d, --debug                 Enable debug mode

Help Options:
  -h, --help                  Show this help message

Available commands:
  config  generate bbsimctl configuration
  olt     OLT Commands
  onus    List ONU Devices
```

`bbsimctl` can be configured via a config file such as:

``` bash
$ cat ~/.bbsim/config
apiVersion: v1
server: 127.0.0.1:50070
grpc:
  timeout: 10s
```

Some example commands:

```bash
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
```

### Autocomplete

`bbsimctl` comes with autocomplete, just run:

```bash
source <(bbsimctl completion bash)
```

## Documentation

More advanced documentation lives in the [here](./docs/README.md)

> This project structure is based on [golang-standards/project-layout](https://github.com/golang-standards/project-layout).
