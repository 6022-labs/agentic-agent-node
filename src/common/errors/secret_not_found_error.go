package errors

import stderrors "errors"

// ErrSecretNotFound is returned by a secret store Read when the key is absent.
var ErrSecretNotFound = stderrors.New("secret not found")
