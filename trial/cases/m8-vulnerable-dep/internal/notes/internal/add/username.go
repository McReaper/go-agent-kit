package add

import "golang.org/x/text/secure/precis"

func Username(s string) (string, error) {
	return precis.UsernameCaseMapped.String(s)
}
