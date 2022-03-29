package main

type asdf struct {
	test string
}

func noParams() {
	return
}

func argl(i string, a, b int) (ret1 string, ret2 interface{}, ret3, ret4 int, ret5 asdf) { // want "named return ret1" "named return ret2" "named return ret3" "named return ret4" "named return ret5"
	return "", nil, 1, 2, asdf{}
}

func good(i string) string {
	return i
}

func myLog(format string, args ...interface{}) {
	return
}
