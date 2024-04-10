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

func GroupByProperty[T any, K comparable](items []T, getProperty func(T) K) map[K][]T {
	grouped := make(map[K][]T)

	for _, item := range items {
		key := getProperty(item)
		grouped[key] = append(grouped[key], item)
	}

	return grouped
}

// Chunk batches []T into [][]T in groups of size. The final chunk of []T will be
// smaller than size if the input slice cannot be chunked evenly. It does not
// make any copies of slice elements.
//
// As an example, take a slice of 5 integers and create chunks of 2 integers
// each (the final value creates a short chunk):
//
//	slices.Chunk([]int{1, 2, 3, 4, 5}, 2) = [][]int{{1, 2}, {3, 4}, {5}}
func Chunk[T any](slice []T, size int) [][]T {
	var chunks [][]T
	for i := 0; i < len(slice); {
		// Clamp the last chunk to the slice bound as necessary.
		end := size
		if l := len(slice[i:]); l < size {
			end = l
		}

		// Set the capacity of each chunk so that appending to a chunk does not
		// modify the original slice.
		chunks = append(chunks, slice[i:i+end:i+end])
		i += end
	}

	return chunks
}
