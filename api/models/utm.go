package models

// This is the glosary object, its saved in an another table
type UtmDefinition struct {
    Id          int          `json:"id" db:"id"`
    Code        string       `json:"code" db:"code"`
    Source      NullString   `json:"utm_source" db:"utm_source"`
    Medium      NullString   `json:"utm_medium" db:"utm_medium"`
    Campaign    NullString   `json:"utm_campaign" db:"utm_campaign"`
    Ad          NullString   `json:"utm_ad" db:"utm_ad"`
    Channel     NullString   `json:"utm_channel" db:"utm_channel"`
}

// In the DB its saved in the communication table
type Utm struct {
    Source      NullString   `json:"utm_source" db:"utm_source"`
    Medium      NullString   `json:"utm_medium" db:"utm_medium"`
    Campaign    NullString   `json:"utm_campaign" db:"utm_campaign"`
    Ad          NullString   `json:"utm_ad" db:"utm_ad"`
    Channel     NullString   `json:"utm_channel" db:"utm_channel"`
}

