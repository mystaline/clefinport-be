package dto

type PaginationResult[T any] struct {
	TotalRecords int `json:"totalRecords" bson:"totalRecords"`
	Data         []T `json:"data"         bson:"data"`
}

type SetSoftDelete struct {
	IsDeleted bool   `column:"is_deleted"`
	DeletedAt string `column:"deleted_at"`
}
