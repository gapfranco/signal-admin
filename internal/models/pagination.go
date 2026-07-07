package models

type PaginationMetadata struct {
	CurrentPage int
	PageSize    int
	TotalItems  int
	TotalPages  int
	HasPrev     bool
	HasNext     bool
	PrevPage    int
	NextPage    int
}
