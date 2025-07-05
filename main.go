package main

func main() {
	const dbFile = "db/databse.sqlite"
	_ = InitServer(dbFile)
	InitAPIRoute()
}
