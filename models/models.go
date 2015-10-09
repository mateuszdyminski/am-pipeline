package models

// User holds info about AM user.
type User struct {
	Pnum      int64    `json:"id,omitempty"`
	Email     string   `json:"email,omitempty"`
	Dob       *string  `json:"dob,omitempty"`
	Weight    *int     `json:"weight,omitempty"`
	Height    *int     `json:"height,omitempty"`
	Nickname  *string  `json:"nickname,omitempty"`
	Country   int      `json:"country,omitempty"`
	City      *string  `json:"city,omitempty"`
	Caption   *string  `json:"caption,omitempty"`
	Longitude *string  `json:"longitude,omitempty"`
	Latitude  *string  `json:"latitude,omitempty"`
	Gender    *int     `json:"gender,omitempty"`
	Score     *float64 `json:"score,omitempty"`
}
