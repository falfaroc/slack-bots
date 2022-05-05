package util

const PersistentFile = "persistent-pies.json"

type PieContent struct {
	Members []Member `json:"members"`
	Pies    []Pie    `json:"pies"`
	State   State    `json:"state"`
}

type Member struct {
	ID string `json:"id"`
}

type Pie struct {
	Type   string `json:"type"`
	Date   string `json:"date"`
	Member Member `json:"member"`
}

type State struct {
	Current  Order `json:"current"`
	Next     Order `json:"next"`
	Previous Order `json:"previous"`
}

type Order struct {
	Date string `json:"date"`
	ID   string `json:"id"`
}
