package model

type FilterResponse struct {
	Results []*User `json:"results"`
	Count   int     `json:"count"`
}
