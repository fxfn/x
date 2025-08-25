package auth

import "encoding/json"

type Server struct {
	AuthorizationEndpoint                              string   `json:"authorization_endpoint"`
	DeviceAuthorizationEndpoint                        string   `json:"device_authorization_endpoint"`
	ClaimsParameterSupported                           bool     `json:"claims_parameter_supported"`
	CodeChallengeMethodsSupported                      []string `json:"code_challenge_methods_supported"`
	EndSessionEndpoint                                 string   `json:"end_session_endpoint"`
	GrantTypesSupported                                []string `json:"grant_types_supported"`
	IdTokenSigningAlgValuesSupported                   []string `json:"id_token_signing_alg_values_supported"`
	Issuer                                             string   `json:"issuer"`
	JwksUri                                            string   `json:"jwks_uri"`
	ResponseModesSupported                             []string `json:"response_modes_supported"`
	ResponseTypesSupported                             []string `json:"response_types_supported"`
	TokenEndpointAuthMethodsSupported                  []string `json:"token_endpoint_auth_methods_supported"`
	TokenEndpointAuthSigningAlgValuesSupported         []string `json:"token_endpoint_auth_signing_alg_values_supported"`
	TokenEndpoint                                      string   `json:"token_endpoint"`
	RequestObjectSigningAlgValuesSupported             []string `json:"request_object_signing_alg_values_supported"`
	UserinfoEndpoint                                   string   `json:"userinfo_endpoint"`
	UserinfoSigningALgValuesSupported                  []string `json:"userinfo_signing_alg_values_supported"`
	IntrospectionEndpoint                              string   `json:"introspection_endpoint"`
	IntrospectionEndpointAuthMethodsSupported          []string `json:"introspection_endpoint_auth_methods_supported"`
	IntrospectionEndpointAuthSigningAlgValuesSupported []string `json:"introspection_endpoint_auth_signing_alg_values_supported"`
	RevocationEndpoint                                 string   `json:"revocation_endpoint"`
	RevocationEndpointAuthMethodsSupported             []string `json:"revocation_endpoint_auth_methods_supported"`
	RevocationEndpointAuthSigningAlgValuesSupported    []string `json:"revocation_endpoint_auth_signing_alg_values_supported"`
}

func NewServer(metadata map[string]any) (*Server, error) {
	serverJson, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}

	var server Server
	err = json.Unmarshal(serverJson, &server)
	if err != nil {
		return nil, err
	}

	return &server, nil
}
