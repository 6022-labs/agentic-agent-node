package requests

type AuthorizeRequest struct {
	WalletAddressRequest
}

func (r *AuthorizeRequest) ValidateAndSanitize() error {
	return r.WalletAddressRequest.ValidateAndSanitize()
}
