package main

//学习闭包

import "fmt"

func main() {
	test()
}

func test() {
	//addTest()
	f1Test()
}

func addTest() {
	tmp1 := add(10)               //10没有被销毁
	fmt.Println(tmp1(1), tmp1(2)) //11 13
	// 此时tmp1和tmp2不是一个实体了
	tmp2 := add(100)
	fmt.Println(tmp2(1), tmp2(2))
}

func add(base int) func(int) int {
	return func(i int) int {
		base += i
		return base
	}
}

func f1Test() {
	fun := f1(100)
	result := fun(11)
	fmt.Println(result)
}

func f1(num int) (returnA func(rAParam int) int) {
	return func(b int) int {
		return num + b
	}
}
