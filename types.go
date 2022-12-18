package main

type Database struct {
	BotUpdateOffset int
	Products        []*Product
}

type Product struct {
	ChatID int64
	URL    string
	Price  string
}
