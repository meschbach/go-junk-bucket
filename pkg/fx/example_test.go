package fx

import "fmt"

func ExampleFilter() {
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	evens := Filter(numbers, func(n int) bool {
		return n%2 == 0
	})

	fmt.Println(evens)
	// Output: [2 4 6 8 10]
}

func ExampleMap() {
	numbers := []int{1, 2, 3}

	doubled := Map(numbers, func(n int) int {
		return n * 2
	})

	fmt.Println(doubled)
	// Output: [2 4 6]
}

func ExampleMap_types() {
	numbers := []int{10, 20, 30}

	labels := Map(numbers, func(n int) string {
		return fmt.Sprintf("%dpx", n)
	})

	fmt.Println(labels)
	// Output: [10px 20px 30px]
}

func ExampleSplit() {
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	under, over := Split(numbers, func(n int) bool {
		return n <= 5
	})

	fmt.Println("≤5:", under)
	fmt.Println(">5:", over)
	// Output:
	// ≤5: [1 2 3 4 5]
	// >5: [6 7 8 9 10]
}

func ExampleIdentity() {
	x := Identity(42)
	fmt.Println(x)

	s := Identity("hello")
	fmt.Println(s)
	// Output:
	// 42
	// hello
}
