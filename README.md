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
cd deployments/helm-chart
helm install -n bbsim bbsim
```

### OLT Provisioning

Once `BBSim` is up and running you can provision the OLT in VOLTHA.

When you install the `BBSim` helm chart you'll notice that the last line of the output
prints the service name and port:

```
NOTES:
BBSim deployed with release name: bbsim

OLT ID: 0
# of NNI Ports: 1
# of PON Ports: 1
# of ONU Ports: 1
Total ONUs: (total: 1)

OLT is listening on: "voltha.svc.bbsim-olt-id-0:50060"
```

#### VOLTHA 1.X

Connect to the voltha CLI and execute this commands:

```
preprovision_olt -t openolt -H voltha.svc.bbsim-olt-id-0:50060
enable
```

## Control API

BBSim comes with a gRPC interface to control the internal state.
We plan to provide a `bbsimctl` at certain point, meanwhile you can use `grpcurl`:

```
$ export BBSIM_IP="$(kubectl get svc -n voltha bbsim-olt-id-0 -o go-template='{{.spec.clusterIP}}')"
$ grpcurl -plaintext $BBSIM_IP:50070 bbsim.BBSim/Version
{
  "version": "0.0.1-alpha",
  "buildTime": "”2019.08.09.084157”",
  "commitHash": "9ef7241b07a83c326ef152320428f204f7eff43d"
}


$ grpcurl -plaintext $BBSIM_IP:50070 bbsim.BBSim/GetOlt
{
  "ID": 22,
  "OperState": "up",
  "NNIPorts": [
    {
      "OperState": "down"
    }
  ],
  "PONPorts": [
    {
      "OperState": "down"
    }
  ]
}
```

## Development

To use a patched version of the `omci-sim` library:

```bash
make dep
cd vendor/github.com/opencord/
rm -rf omci-sim/
git clone https://gerrit.opencord.org/omci-sim
cd omci-sim
```

Once done, go to `gerrit.opencord.org` and locate the patch you want to get. Click on the download URL and copy the `Checkout` command.

It should look something like:

```
git fetch ssh://teone@gerrit.opencord.org:29418/omci-sim refs/changes/67/15067/1 && git checkout FETCH_HEAD
```

Then just execute that command in the `omci-sim` folder inside the vendored dependencies.

> This project structure is based on [golang-standards/project-layout](https://github.com/golang-standards/project-layout).
