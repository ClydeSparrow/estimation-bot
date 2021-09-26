package common

type Person struct {
	Name string
	ID   int
}

type Data struct {
	Key     string
	Author  Person
	Message string
}
