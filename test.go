package main

import (
	"fmt"
	"time"
)

func main() {
	//// 测试数组
	//nums := []int{3, 4, 1, 5, 9, 7, 8, 6}
	//fmt.Println("原始数组:", nums)
	//
	//// 测试快速排序
	//quickSort(nums)
	//fmt.Println("快速排序结果:", nums)
	//heapSortMax(nums)
	//
	//fmt.Println("大顶堆排序结果", nums)
	//heapSortMin(nums)
	//fmt.Println("小顶堆排序结果", nums)
	////testMergeSortArr := []int{3, 4, 1, 5, 9, 7, 8, 6}
	////resSort := MergeSort(testMergeSortArr)
	////fmt.Println(resSort)

	nums := []int{3, 2, 1, 5, 6, 4}
	findKthLargest(nums, 2)
	fmt.Println(nums)
	time.Sleep(2 * time.Second)
	ch1 := make(chan int)
	ch2 := make(chan int)

	go func() {
		for i := 0; i < 5; i++ {
			ch1 <- i
		}
		close(ch1)
	}()
	go func() {
		for {
			if res, ok := <-ch1; ok {
				ch2 <- res * res
			} else {
				break
			}
		}
		close(ch2)
	}()
	for i := range ch2 {
		fmt.Println(i)
	}

}

func findKthLargest(nums []int, k int) int {
	n := len(nums)
	var quickSort func(l, r int) int
	var partition func(l, r int) int
	quickSort = func(l, r int) int {
		pos := partition(l, r)
		if pos == n-k {
			return pos
		} else if pos < n-k {
			return quickSort(pos+1, r)
		} else {
			return quickSort(l, pos-1)
		}
	}
	partition = func(l, r int) int {
		pivot := nums[l]
		i, j := l+1, r
		for {
			for i <= j && nums[i] < pivot {
				i++
			}
			for i <= j && nums[j] > pivot {
				j--
			}
			if i > j {
				break
			}
			nums[i], nums[j] = nums[j], nums[i]
			i++
			j--
		}
		nums[l], nums[j] = nums[j], nums[l]
		return j

	}
	return quickSort(0, n-1)
}
