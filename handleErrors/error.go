package handleerrors

import "errors"

var InternalErrors error = errors.New("Internal server errors")