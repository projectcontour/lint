// Copyright Project Contour Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package a

import (
	// envoy_api_v2_auth "github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
	envoy_config_filter_http_ext_authz_v2 "github.com/envoyproxy/go-control-plane/envoy/config/filter/http/ext_authz/v2"
	contour_api_v1 "github.com/projectcontour/contour/apis/projectcontour/v1"
	contour_api_v1alpha1 "github.com/projectcontour/contour/apis/projectcontour/v1alpha1"
	kingpin_v2 "gopkg.in/alecthomas/kingpin.v2"
	api_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	api_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gateway_v1alpha1 "sigs.k8s.io/gateway-api/apis/v1alpha1"
	gatewayapi_v1alpha1 "sigs.k8s.io/gateway-api/apis/v1alpha1"
)

func foo() {
	meta_v1.Now()
	api_meta_v1.Now()
	api_v1.Now()
	// _ = envoy_api_v2_auth.CertificateValidationContext_ACCEPT_UNTRUSTED
	_ = envoy_config_filter_http_ext_authz_v2.AuthorizationRequest{}
	contour_api_v1.AddKnownTypes(nil)
	_ = contour_api_v1alpha1.GroupVersion
	kingpin_v2.Parse()
	_ = gatewayapi_v1alpha1.GroupVersion
	_ = gateway_v1alpha1.GroupVersion
}
