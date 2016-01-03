package transform

func check(errs ...error) {
	for _, err := range errs {
		if err != nil {
			panic(err)
		}
	}
}
