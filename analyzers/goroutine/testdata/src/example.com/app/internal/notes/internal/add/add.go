package add

func Run(f func()) {
	go f() // want `M11: a goroutine started here has no owner`
}
