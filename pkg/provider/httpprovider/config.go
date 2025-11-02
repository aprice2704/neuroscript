// NeuroScript Version: 0.8.0
// File version: 2
// Purpose: Corrected to parse the 'generic_http' block from the provider params map.
// filename: pkg/provider/httpprovider/config.go
// nlines: 70
// risk_rating: MEDIUM

package httpprovider

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
)

const configKey = "generic_http"

// httpProviderConfig holds the data-driven configuration for a generic HTTP request.
// This struct is populated from the AgentModel.Params["generic_http"] map.
type httpProviderConfig struct {
	// URL is the full HTTP endpoint to call.
	URL string `mapstructure:"api_url"`
	// Headers is a map of HTTP headers to send.
	// Values can contain interpolation tokens like {API_KEY}.
	Headers map[string]string `mapstructure:"api_headers"`
	// BodyTemplate is the JSON request body.
	// It can be a map[string]interface{} or a string.
	// It can contain interpolation tokens like {MODEL} and {PROMPT}.
	BodyTemplate any `mapstructure:"api_body_template"`
	// ResponsePath is a JMESPath/JSONPath string to extract the response text.
	// e.g., "choices[0].message.content"
	ResponsePath string `mapstructure:"api_response_path"`
	// ErrorPath is a JMESPath/JSONPath string to extract an error message from
	// a non-200 response body. e.g., "error.message"
	ErrorPath string `mapstructure:"api_error_path"`
}

// extractConfig parses the httpProviderConfig from the provider params map.
// This is the FIX for "model.Params undefined".
func extractConfig(params map[string]any, agentModelName string) (*httpProviderConfig, error) {
	configMap, ok := params[configKey]
	if !ok {
		return nil, fmt.Errorf("%w: AgentModel '%s' is missing required '%s' params block",
			ErrConfigMissing, agentModelName, configKey)
	}

	var config httpProviderConfig
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           &config,
		WeaklyTypedInput: true,
		TagName:          "mapstructure",
	})
	if err != nil {
		return nil, fmt.Errorf("internal error creating decoder: %w", err)
	}

	if err := decoder.Decode(configMap); err != nil {
		return nil, fmt.Errorf("%w: failed to decode '%s' block: %v",
			ErrConfigInvalid, configKey, err)
	}

	// Validate required fields
	if config.URL == "" {
		return nil, fmt.Errorf("%w: '%s.api_url' is required", ErrConfigInvalid, configKey)
	}
	if config.BodyTemplate == nil {
		return nil, fmt.Errorf("%w: '%s.api_body_template' is required", ErrConfigInvalid, configKey)
	}
	if config.ResponsePath == "" {
		return nil, fmt.Errorf("%w: '%s.api_response_path' is required", ErrConfigInvalid, configKey)
	}

	return &config, nil
}
