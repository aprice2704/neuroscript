// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Defines specifications for hashing and UUID tools.
// filename: pkg/tool/crypto/tooldefs_crypto_hash.go
// nlines: 55
// risk_rating: MEDIUM

package crypto

import (
	"github.com/aprice2704/neuroscript/pkg/policy/capability"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

var cryptoHashToolsToRegister = []tool.ToolImplementation{
	{
		Spec: tool.ToolSpec{
			Name:        "Hash",
			Group:       group,
			Description: "Computes a cryptographic hash of a string. Requires 'crypto:use:hash' capability.",
			Category:    "Cryptography",
			Args: []tool.ArgSpec{
				{Name: "input_string", Type: tool.ArgTypeString, Required: true, Description: "The string to hash."},
				{Name: "algorithm", Type: tool.ArgTypeString, Required: true, Description: "The hash algorithm (e.g., 'sha256', 'sha512')."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns the hex-encoded hash of the input string.",
			Example:         `crypto.Hash(input_string: "hello", algorithm: "sha256")`,
			ErrorConditions: "ErrInvalidArgument if an unsupported algorithm is specified.",
		},
		Func:         toolHash,
		RequiredCaps: []capability.Capability{capability.New(group, capability.VerbUse, "hash")},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "HMAC",
			Group:       group,
			Description: "Computes a Keyed-Hash Message Authentication Code (HMAC). Requires 'crypto:use:hash' capability.",
			Category:    "Cryptography",
			Args: []tool.ArgSpec{
				{Name: "input_string", Type: tool.ArgTypeString, Required: true, Description: "The message to authenticate."},
				{Name: "secret_key", Type: tool.ArgTypeString, Required: true, Description: "The secret key."},
				{Name: "algorithm", Type: tool.ArgTypeString, Required: true, Description: "The hash algorithm (e.g., 'sha256')."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns the hex-encoded HMAC signature.",
			Example:         `crypto.HMAC(input_string: "msg", secret_key: "key", algorithm: "sha256")`,
			ErrorConditions: "ErrInvalidArgument if an unsupported algorithm is specified.",
		},
		Func:         toolHMAC,
		RequiredCaps: []capability.Capability{capability.New(group, capability.VerbUse, "hash")},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "UUID",
			Group:       group,
			Description: "Generates a new, random Version 4 UUID.",
			Category:    "Cryptography",
			Args:        []tool.ArgSpec{},
			ReturnType:  tool.ArgTypeString,
			ReturnHelp:  "Returns a new UUID as a string.",
			Example:     `crypto.UUID()`,
		},
		Func: toolUUID,
	},
}
