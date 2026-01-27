package dto

// PaginationRequest represents pagination parameters for list endpoints
type PaginationRequest struct {
	Limit  int `json:"limit,omitempty" default:"50"`
	Offset int `json:"offset,omitempty" default:"0"`
}

// PaginationMeta represents pagination metadata in responses
type PaginationMeta struct {
	Total   int  `json:"total"`
	Limit   int  `json:"limit"`
	Offset  int  `json:"offset"`
	HasMore bool `json:"has_more"`
}

// GetLimit returns pagination limit, ensuring it's within bounds (1-1000)
func (p *PaginationRequest) GetLimit() int {
	if p.Limit <= 0 || p.Limit > 1000 {
		return 50 // default
	}
	return p.Limit
}

// GetOffset returns pagination offset, ensuring it's >= 0
func (p *PaginationRequest) GetOffset() int {
	if p.Offset < 0 {
		return 0
	}
	return p.Offset
}
