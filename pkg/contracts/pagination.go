package contracts

type Pagination interface {
	GetPagination() (limit int64, skip int64)
}
