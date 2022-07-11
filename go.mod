module github.com/opencord/bbsim

go 1.12

replace (
	github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
	google.golang.org/grpc => google.golang.org/grpc v1.25.1
)

require (
	github.com/Shopify/sarama v1.26.1
	github.com/boguslaw-wojcik/crc32a v1.0.0
	github.com/golang/protobuf v1.5.2
	github.com/google/gopacket v1.1.17
	github.com/google/uuid v1.1.2
	github.com/gorilla/mux v1.7.3
	github.com/grpc-ecosystem/grpc-gateway v1.12.2
	github.com/imdario/mergo v0.3.11
	github.com/jessevdk/go-flags v1.4.0
	github.com/jpillora/backoff v1.0.0
	github.com/looplab/fsm v0.1.0
	github.com/olekukonko/tablewriter v0.0.4
	github.com/opencord/cordctl v0.0.0-20190909161711-01e9c1f04bf4
	github.com/opencord/device-management-interface v1.4.0
	github.com/opencord/omci-lib-go/v2 v2.2.1
	github.com/opencord/voltha-protos/v5 v5.2.3
	github.com/pkg/errors v0.8.1 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.5.1
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/genproto v0.0.0-20220208230804-65c12eb4c068
	google.golang.org/grpc v1.44.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/yaml.v2 v2.3.0
	gotest.tools v2.2.0+incompatible
)
