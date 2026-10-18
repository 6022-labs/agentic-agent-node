package requests

type ListConversationsRequest struct {
	Page            *uint `query:"page"            json:"page,omitempty"`
	ElementsPerPage *uint `query:"elementsPerPage" json:"elementsPerPage,omitempty"`
}

// ValidateAndSanitize clamps paging instead of refusing it.
func (r *ListConversationsRequest) ValidateAndSanitize() error {
	if r.Page == nil || *r.Page < 1 {
		defaultPage := uint(1)
		r.Page = &defaultPage
	}

	if r.ElementsPerPage == nil || *r.ElementsPerPage < 1 {
		defaultElementsPerPage := uint(10)
		r.ElementsPerPage = &defaultElementsPerPage
	}

	if *r.ElementsPerPage > uint(100) {
		maxElementsPerPage := uint(100)
		r.ElementsPerPage = &maxElementsPerPage
	}

	return nil
}
