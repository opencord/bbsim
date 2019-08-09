# BroadBand Simular (BBSim) 

## Build

BBSim is managed via a `Makefile`, plese run the following command
to display all the available targets

```
make help
```

## Usage

## Control API

BBSim comes with a gRPC interface to control the internal state.
We plan to provide a `bbsimctl` at certain point, meanwhile you can use `grpcurl`:

```
$ grpcurl -plaintext 0.0.0.0:50070 bbsim.BBSim/Version
{
  "version": "0.0.1-alpha",
  "buildTime": "”2019.08.09.084157”",
  "commitHash": "9ef7241b07a83c326ef152320428f204f7eff43d"
}


$ grpcurl -plaintext 0.0.0.0:50070 bbsim.BBSim/GetOlt
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

> This project strucutre is based on [golang-standards/project-layout](https://github.com/golang-standards/project-layout).