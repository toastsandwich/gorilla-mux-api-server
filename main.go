package main

import "log"

func main() {
	// fmt.Println("Hello world")
	store, err := NewMySqlStore()
	if err != nil {
		log.Panic(err)
	}
	err = store.Init()
	if err != nil {
		log.Fatal(err)
	}
	app := NewAPIServer(":8080", store)
	log.Fatal(app.run())
}
