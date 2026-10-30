package run

func Start(f func()) {
	go f()
}
