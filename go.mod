module storj.io/uplink

go 1.17

require (
	github.com/minio/highwayhash v1.0.2
	github.com/pkg/profile v1.6.0
	github.com/spacemonkeygo/monkit/v3 v3.0.17
	github.com/stretchr/testify v1.7.0
	github.com/vivint/infectious v0.0.0-20200605153912-25a574ae18a3
	github.com/zeebo/blake3 v0.2.2
	github.com/zeebo/errs v1.2.2
	github.com/zeebo/xxh3 v1.0.1
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	storj.io/common v0.0.0-20220131120956-e74f624a3d55
	storj.io/drpc v0.0.29
)

require (
	github.com/calebcase/tmpfile v1.0.3 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/pprof v0.0.0-20211108044417-e9b028704de0 // indirect
	github.com/klauspost/cpuid/v2 v2.0.9 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/zeebo/admission/v3 v3.0.3 // indirect
	github.com/zeebo/float16 v0.1.0 // indirect
	github.com/zeebo/incenc v0.0.0-20180505221441-0d92902eec54 // indirect
	golang.org/x/crypto v0.0.0-20211108221036-ceb1ce70b4fa // indirect
	golang.org/x/sys v0.0.0-20220209214540-3681064d5158 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c // indirect
)

replace storj.io/common => ../common
