module gerrit.opencord.org/bbsim

go 1.12

require (
	github.com/golang/protobuf v1.3.2
	github.com/looplab/fsm v0.1.0
	github.com/opencord/omci-sim v0.0.0-20190717165025-5ff7bb17f1e9
	github.com/opencord/voltha-protos v0.0.0-20190813191205-792553b747df
	github.com/sirupsen/logrus v1.4.2
	google.golang.org/grpc v1.22.1
)

replace github.com/opencord/omci-sim v0.0.0-20190717165025-5ff7bb17f1e9 => ./vendor/github.com/opencord/omci-sim