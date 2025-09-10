// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Defines specifications for JWT (JSON Web Token) tools.
// filename: pkg/tool/crypto/tooldefs_crypto.go
// nlines: 83
// risk_rating: HIGH

package crypto

import (
	"github.com/aprice2704/neuroscript/pkg/policy/capability"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

const group = "crypto"

var cryptoToolsToRegister = []tool.ToolImplementation{
	{
		Spec: tool.ToolSpec{
			Name:        "SignJWT",
			Group:       group,
			Description: "Creates and signs a JWT token using a specified HMAC algorithm and secret. Requires 'crypto:sign:jwt' capability.",
			Category:    "Cryptography",
			Args: []tool.ArgSpec{
				{Name: "claims_map", Type: tool.ArgTypeMap, Required: true, Description: "A map of claims for the token payload."},
				{Name: "secret", Type: tool.ArgTypeString, Required: true, Description: "The shared secret key for signing."},
				{Name: "algorithm", Type: tool.ArgTypeString, Required: true, Description: "The signing algorithm (e.g., 'HS256', 'HS384', 'HS512')."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns the signed JWT as a string.",
			Example:         `crypto.SignJWT(claims_map: {"sub": "user123"}, secret: "your-secret", algorithm: "HS256")`,
			ErrorConditions: "ErrArgumentMismatch for wrong arg count/types. ErrInvalidArgument for unsupported algorithm.",
		},
		Func:          toolSignJWT,
		RequiresTrust: true,
		RequiredCaps: []capability.Capability{
			capability.New(capability.ResCrypto, capability.VerbSign, "jwt"),
		},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "VerifyJWT",
			Group:       group,
			Description: "Verifies a JWT's signature and claims (like expiry) against a secret.",
			Category:    "Cryptography",
			Args: []tool.ArgSpec{
				{Name: "token_string", Type: tool.ArgTypeString, Required: true, Description: "The JWT to verify."},
				{Name: "secret", Type: tool.ArgTypeString, Required: true, Description: "The shared secret key for verification."},
			},
			ReturnType:      tool.ArgTypeMap,
			ReturnHelp:      "Returns a map of the claims if the token is valid. Throws an error if verification fails.",
			Example:         `crypto.VerifyJWT(token_string: "ey...", secret: "your-secret")`,
			ErrorConditions: "ErrArgumentMismatch for wrong arg count/types. ErrInvalidArgument if token is invalid or verification fails.",
		},
		Func: toolVerifyJWT,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "DecodeJWT",
			Group:       group,
			Description: "Decodes a JWT and returns its claims without verifying the signature.",
			Category:    "Cryptography",
			Args: []tool.ArgSpec{
				{Name: "token_string", Type: tool.ArgTypeString, Required: true, Description: "The JWT to decode."},
			},
			ReturnType:      tool.ArgTypeMap,
			ReturnHelp:      "Returns a map containing the token's claims.",
			Example:         `crypto.DecodeJWT(token_string: "ey...")`,
			ErrorConditions: "ErrArgumentMismatch for wrong arg count/type. ErrInvalidArgument if the token is malformed.",
		},
		Func: toolDecodeJWT,
	},
}
