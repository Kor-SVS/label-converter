package lab

func TernaryOperator[T any](state bool, t, f T) T {
	if state {
		return t
	} else {
		return f
	}
}
