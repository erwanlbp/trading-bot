package util

type (
	Mapper[T, K any] func(T) K
	Predicate[T any] Mapper[T, bool]
)

func Keys[K comparable, V any](m map[K]V) []K {
	var res []K
	for k := range m {
		res = append(res, k)
	}
	return res
}

func AsMap[T any, K comparable](slice []T, keyGetter Mapper[T, K]) map[K]T {
	res := make(map[K]T)
	for _, e := range slice {
		res[keyGetter(e)] = e
	}
	return res
}

func Exists[T any](slice []T, finder Predicate[T]) bool {
	for _, e := range slice {
		if finder(e) {
			return true
		}
	}
	return false
}

func Identity[T any](t T) T {
	return t
}

func Map[T, V any](slice []T, mapper Mapper[T, V]) []V {
	var res []V = make([]V, len(slice))
	for i, e := range slice {
		res[i] = mapper(e)
	}
	return res
}

func FilterSlice[T any](slice []T, predicate Predicate[T]) []T {
	var res []T
	for _, t := range slice {
		if predicate(t) {
			res = append(res, t)
		}
	}
	return res
}
