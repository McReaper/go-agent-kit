package m2case

import "errors"

const Greeting = "we recieve notes"

func Get_note() string {
	return Greeting
}

func Copy(len int) int {
	return len
}

func Fail() error {
	return errors.New("Bad thing.")
}
