// NeuroScript Version: 0.5.2
// File version: 2
// Purpose: Implements JWT (JSON Web Token) creation, verification, and decoding tools. Corrected error handling in DecodeJWT.
// filename: pkg/tool/crypto/tools_crypto.go
// nlines: 109
// risk_rating: HIGH

package crypto

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/golang-jwt/jwt/v5"
)

func toolSignJWT(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 3 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "SignJWT: expected 3 arguments (claims_map, secret, algorithm)", lang.ErrArgumentMismatch)
	}

	claimsMap, ok1 := args[0].(map[string]interface{})
	secret, ok2 := args[1].(string)
	algo, ok3 := args[2].(string)

	if !ok1 || !ok2 || !ok3 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "SignJWT: invalid argument types", lang.ErrArgumentMismatch)
	}

	signingMethod := jwt.GetSigningMethod(algo)
	if signingMethod == nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInvalidValue, fmt.Sprintf("SignJWT: unsupported algorithm '%s'", algo), lang.ErrInvalidArgument)
	}

	token := jwt.NewWithClaims(signingMethod, jwt.MapClaims(claimsMap))
	signedString, err := token.SignedString([]byte(secret))
	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, fmt.Sprintf("SignJWT: failed to sign token: %v", err), lang.ErrInternal)
	}

	// interpreter.GetLogger().Debug("Tool: SignJWT", "algorithm", algo)
	return signedString, nil
}

func toolVerifyJWT(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "VerifyJWT: expected 2 arguments (token_string, secret)", lang.ErrArgumentMismatch)
	}

	tokenString, ok1 := args[0].(string)
	secret, ok2 := args[1].(string)
	if !ok1 || !ok2 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "VerifyJWT: invalid argument types", lang.ErrArgumentMismatch)
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInvalidValue, fmt.Sprintf("VerifyJWT: token verification failed: %v", err), lang.ErrInvalidArgument)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// interpreter.GetLogger().Debug("Tool: VerifyJWT", "claims", claims)
		return (map[string]interface{})(claims), nil
	}

	return nil, lang.NewRuntimeError(lang.ErrorCodeInvalidValue, "VerifyJWT: token is invalid", lang.ErrInvalidArgument)
}

func toolDecodeJWT(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "DecodeJWT: expected 1 argument (token_string)", lang.ErrArgumentMismatch)
	}
	tokenString, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, "DecodeJWT: token_string must be a string", lang.ErrArgumentMismatch)
	}

	// Note: ParseUnverified is for decoding without signature validation.
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		// Any error from ParseUnverified indicates a malformed token.
		return nil, lang.NewRuntimeError(lang.ErrorCodeInvalidValue, fmt.Sprintf("DecodeJWT: malformed token: %v", err), lang.ErrInvalidArgument)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		// interpreter.GetLogger().Debug("Tool: DecodeJWT", "claims", claims)
		return (map[string]interface{})(claims), nil
	}

	return nil, lang.NewRuntimeError(lang.ErrorCodeInvalidValue, "DecodeJWT: could not parse claims", lang.ErrInvalidArgument)
}
