package add

func AddDefaults(a Args) Args {
	if a.Body == "" {
		a.Body = "-"
	}
	return a
}
