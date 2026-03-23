module smart-slowquery

go 1.22

require (
	git.garena.com/shopee/go-shopeelib v0.1.158
	git.garena.com/shopee/platform/space-sdk v0.3.50
	github.com/Shopify/sarama v1.38.1
	github.com/allegro/bigcache v1.2.1
	github.com/emersion/go-sasl v0.0.0-20231106173351-e73c9f7bad43
	github.com/emersion/go-smtp v0.18.1
	github.com/gin-contrib/pprof v1.4.0
	github.com/gin-gonic/gin v1.9.1
	github.com/go-logr/logr v1.4.1
	github.com/go-sql-driver/mysql v1.7.1
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/jinzhu/configor v1.2.1
	github.com/jinzhu/gorm v1.9.12
	github.com/juju/errors v1.0.0
	github.com/mitchellh/mapstructure v1.5.0
	github.com/panjf2000/ants v1.3.0
	//github.com/pingcap/log v1.1.0 // indirect
	//github.com/pingcap/parser v3.1.2+incompatibl2
	github.com/pingcap/parser v0.0.0-20210525032559-c37778aff307
	//github.com/pingcap/tidb v0.0.0-20190108123336-c68ee7318319
	github.com/pingcap/tidb v1.1.0-beta.0.20210601085537-5d7c852770eb
	github.com/pingcap/tidb/parser v0.0.0-20230623153413-fd2cd685bb9b
	github.com/pingcap/tipb v0.0.0-20210601083426-79a378b6d1c4 // indirect
	github.com/pkg/errors v0.9.1
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.9.3
	github.com/stretchr/testify v1.8.4
	github.com/sykesm/zap-logfmt v0.0.4
	github.com/vjeantet/grok v1.0.1
	github.com/zsais/go-gin-prometheus v0.1.0
	go.uber.org/atomic v1.11.0
	go.uber.org/zap v1.27.0
	golang.org/x/text v0.14.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
	gorm.io/driver/clickhouse v0.6.0
	gorm.io/driver/mysql v1.4.7
	gorm.io/gorm v1.25.2
	sigs.k8s.io/controller-runtime v0.14.6
)

replace github.com/pingcap/log => github.com/pingcap/log v0.0.0-20210317133921-96f4fcab92a4 // indirect

require (
	github.com/ClickHouse/clickhouse-go/v2 v2.17.1
	github.com/CorgiMan/json2 v0.0.0-20150213135156-e72957aba209
	github.com/astaxie/beego v1.12.3
	github.com/dchest/uniuri v1.2.0
	github.com/gedex/inflector v0.0.0-20170307190818-16278e9db813
	github.com/go-playground/assert/v2 v2.2.0
	github.com/juju/ratelimit v1.0.1
	github.com/kr/pretty v0.3.1
	github.com/ory/dockertest/v3 v3.10.0
	github.com/prometheus/client_golang v1.14.0
	github.com/russross/blackfriday v1.6.0
	github.com/saintfish/chardet v0.0.0-20230101081208-5e3ef4b5456d
	github.com/tidwall/gjson v1.12.1
	gopkg.in/resty.v1 v1.12.0
	vitess.io/vitess v0.0.0-20200325000816-eda961851d63
)

require (
	cloud.google.com/go v0.110.4 // indirect
	cloud.google.com/go/iam v1.1.1 // indirect
	cloud.google.com/go/storage v1.30.1 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/containerd/continuity v0.4.3 // indirect
	github.com/denisenkom/go-mssqldb v0.11.0 // indirect
	github.com/dgraph-io/ristretto v0.1.1 // indirect
	github.com/docker/cli v24.0.7+incompatible // indirect
	github.com/docker/docker v24.0.7+incompatible // indirect
	github.com/docker/go-connections v0.5.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/golang/glog v1.1.0 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/googleapis/gax-go/v2 v2.11.0 // indirect
	github.com/imdario/mergo v0.3.13 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lib/pq v1.10.2 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0-rc5 // indirect
	github.com/opencontainers/runc v1.1.10 // indirect
	github.com/pingcap/log v1.1.0 // indirect
	github.com/rogpeppe/go-internal v1.10.0 // indirect
	github.com/shiena/ansicolor v0.0.0-20151119151921-a422bbe96644 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/twmb/murmur3 v1.1.6 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	go.etcd.io/bbolt v1.3.7 // indirect
	golang.org/x/mod v0.14.0 // indirect
	golang.org/x/tools v0.15.0 // indirect
	google.golang.org/api v0.126.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230711160842-782d3b101e98 // indirect
)

require (
	git.garena.com/shopee/common/jsonext v0.1.0 // indirect
	git.garena.com/shopee/platform/splog v1.4.10 // indirect
	git.garena.com/shopee/platform/trace v0.1.0 // indirect
	git.garena.com/shopee/platform/tracing v1.10.4 // indirect
	github.com/BurntSushi/toml v1.2.1 // indirect
	github.com/ClickHouse/ch-go v0.61.3 // indirect
	github.com/andybalholm/brotli v1.1.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bytedance/sonic v1.9.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/chenzhuoyu/base64x v0.0.0-20221115062448-fe3a3abad311 // indirect
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/cznic/mathutil v0.0.0-20181122101859-297441e03548 // indirect
	github.com/danjacques/gofslock v0.0.0-20191023191349-0a45f885bc37 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/eapache/go-resiliency v1.3.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20230111030713-bf00bc1b83b6 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.2 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-faster/city v1.0.1 // indirect
	github.com/go-faster/errors v0.7.1 // indirect
	github.com/go-logr/zapr v1.2.3 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.14.0 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.6.0
	github.com/gotopia/watcher v0.1.3 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/hashicorp/go-version v1.6.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.7.6 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.3 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.17.7 // indirect
	github.com/klauspost/cpuid/v2 v2.2.4 // indirect
	github.com/leodido/go-urn v1.2.4 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/paulmach/orb v0.11.1 // indirect
	github.com/pelletier/go-toml/v2 v2.0.8 // indirect
	github.com/percona/go-mysql v0.0.0-20210427141028-73d29c6da78c
	github.com/pierrec/lz4/v4 v4.1.21 // indirect
	github.com/pingcap/errors v0.11.5-0.20210425183316-da1aaba5fb63 // indirect
	github.com/pingcap/failpoint v0.0.0-20220801062533-2eaa32854a6c // indirect
	github.com/pingcap/kvproto v0.0.0-20210507054410-a8152f8a876c // indirect
	//github.com/pingcap/log v1.1.0 // indirect
	//github.com/pingcap/tipb v0.0.0-20210525032549-b80be13ddf6c // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.39.0 // indirect
	github.com/prometheus/procfs v0.9.0 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/shirou/gopsutil v3.21.11+incompatible // indirect
	github.com/shopspring/decimal v1.3.1 // indirect
	github.com/spf13/afero v1.9.3 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.15.0 // indirect
	github.com/subosito/gotenv v1.4.2 // indirect
	github.com/tikv/pd v1.1.0-beta.0.20210323121136-78679e5e209d // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/uber/jaeger-client-go v2.30.0+incompatible // indirect
	github.com/uber/jaeger-lib v2.4.1+incompatible // indirect
	github.com/ugorji/go/codec v1.2.11 // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	github.com/yusufpapurcu/wmi v1.2.3 // indirect
	go.etcd.io/etcd v0.5.0-alpha.5.0.20200910180754-dd1b699fc489 // indirect
	go.opentelemetry.io/otel v1.24.0 // indirect
	go.opentelemetry.io/otel/trace v1.24.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/arch v0.3.0 // indirect
	golang.org/x/crypto v0.19.0 // indirect
	golang.org/x/exp v0.0.0-20230519143937-03e91628a987 // indirect
	golang.org/x/net v0.21.0 // indirect
	golang.org/x/sync v0.6.0 // indirect
	golang.org/x/sys v0.17.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/grpc v1.58.3 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apimachinery v0.27.4 // indirect
	k8s.io/klog/v2 v2.90.1 // indirect
	k8s.io/utils v0.0.0-20230220204549-a5ecb0141aa5 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
)
