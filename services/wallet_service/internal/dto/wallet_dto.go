package dto

import "time"

type GetWalletInfoResult struct {
	ID             string    `json:"id"`
	FullName       string    `json:"fullName"`
	ProfilePicture *string   `json:"profilePicture"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type GetWalletInfoData struct {
	ID             string    `json:"id"             column:"id::text"`
	FullName       string    `json:"fullName"       column:"full_name"`
	ProfilePicture string    `json:"profilePicture" column:"profile_picture"`
	CreatedAt      time.Time `json:"createdAt"      column:"created_at"`
	UpdatedAt      time.Time `json:"updatedAt"      column:"updated_at"`
}
