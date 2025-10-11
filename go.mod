module github.com/siemens/ghostwire/v2

go 1.24.0

replace github.com/mattn/go-sqlite3 => github.com/mattn/go-sqlite3 v1.14.12

require (
	github.com/cenkalti/backoff/v4 v4.3.0
	github.com/cilium/ebpf v0.19.0
	github.com/containernetworking/cni v1.3.0
	github.com/docker/docker v28.5.1+incompatible
	github.com/dustinkirkland/golang-petname v0.0.0-20240428194347-eebcea082ee0
	github.com/getkin/kin-openapi v0.133.0
	github.com/google/nftables v0.3.0
	github.com/gorilla/mux v1.8.1
	github.com/gorilla/websocket v1.5.3
	github.com/jinzhu/copier v0.4.0
	github.com/kr/pretty v0.3.1
	github.com/mdlayher/ethernet v0.0.0-20220221185849-529eae5b6118
	github.com/mdlayher/genetlink v1.3.2
	github.com/mdlayher/packet v1.1.2
	github.com/ohler55/ojg v1.26.10
	github.com/onsi/ginkgo/v2 v2.26.0
	github.com/onsi/gomega v1.38.2
	github.com/ory/dockertest v3.3.5+incompatible
	github.com/ory/dockertest/v3 v3.12.0
	github.com/siemens/ieddata v1.1.1
	github.com/siemens/mobydig v1.1.0
	github.com/siemens/turtlefinder v1.1.3
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cobra v1.10.1
	github.com/thediveo/deferrer v0.1.0
	github.com/thediveo/enumflag/v2 v2.0.7
	github.com/thediveo/fdooze v0.3.2
	github.com/thediveo/go-plugger/v3 v3.1.1
	github.com/thediveo/ioctl v0.9.4
	github.com/thediveo/lxkns v0.38.1
	github.com/thediveo/morbyd v0.18.0
	github.com/thediveo/namspill v0.1.7
	github.com/thediveo/netdb v1.1.2
	github.com/thediveo/notwork v1.7.1
	github.com/thediveo/nufftables v0.10.0
	github.com/thediveo/osrelease v1.0.3
	github.com/thediveo/procfsroot v1.0.2
	github.com/thediveo/spaserve v1.0.3
	github.com/thediveo/success v1.0.3
	github.com/thediveo/testbasher v1.0.9
	github.com/thediveo/whalewatcher v0.12.0
	github.com/vishvananda/netlink v1.3.1
	golang.org/x/exp v0.0.0-20251009144603-d2f985daa21b
	golang.org/x/sys v0.37.0
	golang.org/x/text v0.30.0
	sigs.k8s.io/kind v0.30.0
)

require (
	al.essio.dev/pkg/shellescape v1.5.1 // indirect
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/containerd/errdefs/pkg v0.3.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/go-viper/mapstructure/v2 v2.1.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mdlayher/devlink v0.0.0-20191111174559-94b7996630ec // indirect
	github.com/moby/go-archive v0.1.0 // indirect
	github.com/moby/sys/userns v0.1.0 // indirect
	github.com/ncruces/go-strftime v0.1.9 // indirect
	github.com/oasdiff/yaml v0.0.0-20250309154309-f31be36b4037 // indirect
	github.com/oasdiff/yaml3 v0.0.0-20250309153720-d2182401db90 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/rogpeppe/go-internal v1.13.1 // indirect
	github.com/thediveo/cpus v0.7.1 // indirect
	github.com/thediveo/faf v0.2.0 // indirect
	github.com/woodsbury/decimal128 v1.3.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.uber.org/automaxprocs v1.6.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	modernc.org/libc v1.65.10 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
	modernc.org/sqlite v1.38.0 // indirect
)

require (
	dario.cat/mergo v1.0.0 // indirect
	github.com/AdaLogics/go-fuzz-headers v0.0.0-20240806141605-e8a1dd7889d6 // indirect
	github.com/AdamKorcz/go-118-fuzz-build v0.0.0-20231105174938-2b5cbb29f3e2 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
	github.com/BurntSushi/toml v1.4.0 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/Microsoft/hcsshim v0.11.7 // indirect
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/containerd/cgroups v1.1.0 // indirect
	github.com/containerd/containerd v1.7.27 // indirect
	github.com/containerd/containerd/api v1.9.0 // indirect
	github.com/containerd/continuity v0.4.5 // indirect
	github.com/containerd/errdefs v1.0.0 // indirect
	github.com/containerd/fifo v1.1.0 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/containerd/platforms v0.2.1 // indirect
	github.com/containerd/ttrpc v1.2.7 // indirect
	github.com/containerd/typeurl/v2 v2.2.3 // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/cli v27.4.1+incompatible // indirect
	github.com/docker/go-connections v0.5.0 // indirect
	github.com/docker/go-events v0.0.0-20190806004212-e31b211e4f1c // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/evanphx/json-patch/v5 v5.6.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/gammazero/deque v0.2.1 // indirect
	github.com/gammazero/workerpool v1.1.3 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20250403155104-27863c87afa6 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gotestyourself/gotestyourself v2.2.0+incompatible // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jmoiron/sqlx v1.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/josharian/native v1.1.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mdlayher/netlink v1.8.0
	github.com/mdlayher/socket v0.5.1 // indirect
	github.com/miekg/dns v1.1.59 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/locker v1.0.1 // indirect
	github.com/moby/patternmatcher v0.6.0 // indirect
	github.com/moby/sys/mountinfo v0.7.2 // indirect
	github.com/moby/sys/sequential v0.6.0 // indirect
	github.com/moby/sys/signal v0.7.0 // indirect
	github.com/moby/sys/user v0.4.0 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/opencontainers/runc v1.2.3 // indirect
	github.com/opencontainers/runtime-spec v1.2.0 // indirect
	github.com/opencontainers/selinux v1.11.0 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/perimeterx/marshmallow v1.1.5 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus-community/pro-bing v0.4.0 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
	github.com/thediveo/caps v0.10.2
	github.com/thediveo/go-mntinfo v1.0.3 // indirect
	github.com/vishvananda/netns v0.0.5 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.47.0 // indirect
	go.opentelemetry.io/otel v1.35.0 // indirect
	go.opentelemetry.io/otel/metric v1.35.0 // indirect
	go.opentelemetry.io/otel/trace v1.35.0 // indirect
	golang.org/x/mod v0.29.0 // indirect
	golang.org/x/net v0.46.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/tools v0.38.0 // indirect
	google.golang.org/genproto v0.0.0-20240123012728-ef4313101c80 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250324211829-b45e905df463 // indirect
	google.golang.org/grpc v1.73.0 // indirect
	google.golang.org/protobuf v1.36.7 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gotest.tools v2.2.0+incompatible // indirect
	k8s.io/cri-api v0.33.2 // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
)
