// NeuroScript Version: 0.4.1
// File version: 1
// Purpose: Provide stub implementations for File API tools to return ErrFeatureNotImplemented.
// filename: pkg/tool/fileapi/tools_file_api_impl.go
// nlines: 32
// risk_rating: LOW

package fileapi

import "github.com/aprice2704/neuroscript/pkg/lang"

// toolListAPIFiles is a stub implementation for the ListAPIFiles tool.
func toolListAPIFiles(interpreter *neurogo.Interpreter, args []interface{}) (interface{}, error) {
	interpreter.Logger().Error("TOOL ListAPIFiles NOT IMPLEMENTED")
	return nil, lang.NewRuntimeError(lang.ErrorCodeFeatureNotImplemented, "Tool 'ListAPIFiles' is not yet implemented.", lang.ErrFeatureNotImplemented)
}

// toolDeleteAPIFile is a stub implementation for the DeleteAPIFile tool.
func toolDeleteAPIFile(interpreter *neurogo.Interpreter, args []interface{}) (interface{}, error) {
	interpreter.Logger().Error("TOOL DeleteAPIFile NOT IMPLEMENTED")
	return nil, lang.NewRuntimeError(lang.ErrorCodeFeatureNotImplemented, "Tool 'DeleteAPIFile' is not yet implemented.", lang.ErrFeatureNotImplemented)
}

// toolUploadFile is a stub implementation for the UploadFile tool.
func toolUploadFile(interpreter *neurogo.Interpreter, args []interface{}) (interface{}, error) {
	interpreter.Logger().Error("TOOL UploadFile NOT IMPLEMENTED")
	return nil, lang.NewRuntimeError(lang.ErrorCodeFeatureNotImplemented, "Tool 'UploadFile' is not yet implemented.", lang.ErrFeatureNotImplemented)
}