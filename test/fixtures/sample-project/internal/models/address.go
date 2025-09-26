package models

type Address struct {
	Street string
	City   string
	Zip    string
}

func (a Address) IsComplete() bool {
	return a.Street != "" && a.City != "" && a.Zip != ""
}
