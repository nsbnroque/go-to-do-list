package models

type D3Response struct {
	Nodes []Node `json:"nodes"`
	Links []Link `json:"links"`
}
