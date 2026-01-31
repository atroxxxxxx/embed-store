package cluster

func squareDistance(vec1, vec2 []float32) float32 {
	var sum float32
	for i := range len(vec1) {
		difference := vec1[i] - vec2[i]
		sum += difference * difference
	}
	return sum
}
