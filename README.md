# lint

Go code linters for Project Contour.

## running
```bash
go install .
go vet -vettool (which lint) ./...
```

## linters

### import alias

consider import path: github.com/projectcontour/x/y/z/v\*

1. the alias name should be subset of `x[optional]_y[optional]_z[optional]_v*` where optional means it can be present or not.
2. one of `x` or `y` or `z` must be present in alias name.
3. If version exists in path, must be specified.
4. words like `apis` should be `api` in import alias

#### Example

**Valid imports**

```go
import meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
import api_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
import api_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
import envoy_api_v2_auth "github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
import envoy_config_filter_http_ext_authz_v2 "github.com/envoyproxy/go-control-plane/envoy/config/filter/http/ext_authz/v2"
import contour_api_v1 "github.com/projectcontour/contour/apis/projectcontour/v1"
import contour_api_v1alpha1 "github.com/projectcontour/contour/apis/projectcontour/v1alpha1"
import kingpin_v2 "gopkg.in/alecthomas/kingpin.v2"
import serviceapis_v1alpha1 "sigs.k8s.io/service-apis/api/v1alpha1"
```

**Invalid imports**

```go
import v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
import meta "k8s.io/apimachinery/pkg/apis/meta/v1"
import meta_api_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
import api_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
```

### messagefmt
linter for log message formatting.

- that logrus log messages are not capitalized
- that kingpin command help is capitalized
