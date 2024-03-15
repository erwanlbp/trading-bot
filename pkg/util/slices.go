package util

type (
	Mapper[T, K any] func(T) K
	Predicate[T any] Mapper[T, bool]
)

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

func Map[T, V any](slice []T, mapper Mapper[T, V]) []V {
	var res []V = make([]V, len(slice))
	for i, e := range slice {
		res[i] = mapper(e)
	}
	return res
}
