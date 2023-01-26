package main

func contains(arr []int, item int) bool {
	for _, v := range arr {
		if v == item {
			return true
		}
	}

	return false
}
