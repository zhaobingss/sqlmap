package engine

import "errors"

var ERR_NOT_GOT_RECORD = errors.New("got record empty")
var ERR_MORE_THAN_ONE_RECORD = errors.New("more than one record")
