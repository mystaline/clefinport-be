package dto

import "time"

type EmbeddedCurrency struct {
	CurrencySymbol string `json:"currencySymbol" column:"profile_settings.currency_symbol"`
	CurrencyName   string `json:"currencyName"   column:"profile_settings.currency_name"`
}

type GetUserInfoResult struct {
	ID             string           `json:"id"`
	FullName       string           `json:"fullName"`
	ProfilePicture *string          `json:"profilePicture"`
	TotalBalance   int              `json:"totalBalance"`
	Timezone       string           `json:"timezone"`
	Currency       EmbeddedCurrency `json:"currency"`
	CreatedAt      time.Time        `json:"createdAt"`
	UpdatedAt      time.Time        `json:"updatedAt"`
}

type GetUserInfoData struct {
	ID             string `json:"id"             column:"users.id::text"`
	FullName       string `json:"fullName"       column:"users.full_name"`
	ProfilePicture string `json:"profilePicture" column:"users.profile_picture"`
	Timezone       string `json:"timezone"       column:"profile_settings.timezone"`
	Currency       EmbeddedCurrency
	CreatedAt      time.Time `json:"createdAt"      column:"users.created_at"`
	UpdatedAt      time.Time `json:"updatedAt"      column:"users.updated_at"`
}
