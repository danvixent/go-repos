package main

//program flags

//search by name
type searchName struct {
	value string
}

//search by description
type searchDesc struct {
	value string
}

//search by creation date
type searchCreationDate struct {
	value string
}

//search language base
type searchLang struct {
	value string
}

func (sn *searchName) Set(s string) {
	sn.value = s
}

func (sn *searchDesc) Set(s string) {
	sn.value = s
}

func (sn *searchCreationDate) Set(s string) {
	sn.value = s
}

func (sn *searchLang) Set(s string) {
	sn.value = s
}

//////////////////////////////////

func (sn *searchName) String(s string) string {

}

func (sn *searchDesc) String(s string) string {
	sn.value = s
}

func (sn *searchCreationDate) String(s string) string {
	sn.value = s
}

func (sn *searchLang) String(s string) string {
	sn.value = s
}

func flags() {

}
