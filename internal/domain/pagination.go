package domain

// PaginationQuery adalah parameter yang dikirim dari Handler ke Usecase/Repo
type PaginationQuery struct {
	Page  int
	Limit int
}

// GetOffset menghitung titik awal data (untuk GORM)
func (p *PaginationQuery) GetOffset() int {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.Limit <= 0 {
		p.Limit = 10 // Default limit
	}
	return (p.Page - 1) * p.Limit
}

// PaginationMeta adalah informasi tambahan yang dikirim balik ke Frontend
type PaginationMeta struct {
	CurrentPage int   `json:"current_page"`
	Limit       int   `json:"limit"`
	TotalItems  int64 `json:"total_items"`
	TotalPages  int   `json:"total_pages"`
}
