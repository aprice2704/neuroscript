// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Implements the tool functions for managing provider accounts.
// filename: pkg/tool/account/tools_account.go
// nlines: 75
// risk_rating: HIGH
package account

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// accountAdminRuntime defines the interface we expect from the runtime
// for account administrative operations.
type accountAdminRuntime interface {
	AccountsAdmin() interfaces.AccountAdmin
	Accounts() interfaces.AccountReader
}

func getAccountAdmin(rt tool.Runtime) (interfaces.AccountAdmin, error) {
	interp, ok := rt.(accountAdminRuntime)
	if !ok {
		return nil, fmt.Errorf("internal error: runtime does not support account admin operations")
	}
	return interp.AccountsAdmin(), nil
}

func getAccountReader(rt tool.Runtime) (interfaces.AccountReader, error) {
	interp, ok := rt.(accountAdminRuntime)
	if !ok {
		return nil, fmt.Errorf("internal error: runtime does not support account read operations")
	}
	return interp.Accounts(), nil
}

func toolRegisterAccount(rt tool.Runtime, args []interface{}) (interface{}, error) {
	admin, err := getAccountAdmin(rt)
	if err != nil {
		return nil, err
	}
	name, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("argument 'name' must be a string")
	}
	config, ok := args[1].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("argument 'config' must be a map[string]interface{}")
	}

	err = admin.Register(name, config)
	if err != nil {
		return nil, err
	}
	return true, nil
}

func toolDeleteAccount(rt tool.Runtime, args []interface{}) (interface{}, error) {
	admin, err := getAccountAdmin(rt)
	if err != nil {
		return nil, err
	}
	name, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("argument 'name' must be a string")
	}
	if ok := admin.Delete(name); !ok {
		return false, nil
	}
	return true, nil
}

func toolListAccounts(rt tool.Runtime, args []interface{}) (interface{}, error) {
	reader, err := getAccountReader(rt)
	if err != nil {
		return nil, err
	}
	names := reader.List()
	return names, nil
}
