package diff

func GetLongestCommonSubsequenceLength(n, m *[]Diffable) uint32 {
	a := *n
	b := *m
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	matrix := make([][]uint32, len(a)+1)
	for i, _ := range matrix {
		matrix[i] = make([]uint32, len(b)+1)
	}
	for i := 0; i <= len(b); i++ {
		matrix[0][i] = 0
	}
	for i := 0; i <= len(a); i++ {
		matrix[i][0] = 0
	}
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			m := max(matrix[i][j-1], matrix[i-1][j])
			if a[i-1].DiffParam() == b[j-1].DiffParam() {
				matrix[i][j] = matrix[i-1][j-1] + 1
			} else {
				matrix[i][j] = m
			}
		}
	}
	return matrix[len(a)][len(b)]
}

func max(x, y uint32) uint32 {
	if x > y {
		return x
	} else {
		return y
	}
}

func GetLongestCommonSubsequence(one, other *[]Diffable) (uint32, *[]Diffable) {
	a := *one
	b := *other
	if len(a) == 0 || len(b) == 0 {
		return 0, &[]Diffable{}
	}
	matrix := make([][]uint32, len(a)+1)
	for i, _ := range matrix {
		matrix[i] = make([]uint32, len(b)+1)
	}
	for i := 0; i <= len(b); i++ {
		matrix[0][i] = 0
	}
	for i := 0; i <= len(a); i++ {
		matrix[i][0] = 0
	}
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			m := max(matrix[i][j-1], matrix[i-1][j])
			if a[i-1].DiffParam() == b[j-1].DiffParam() {
				matrix[i][j] = matrix[i-1][j-1] + 1
			} else {
				matrix[i][j] = m
			}
		}
	}
	lcs := matrix[len(a)][len(b)]
	var s []Diffable
	for i := len(a); i > 0; i-- {
		for j := len(b); j > 0; {
			if matrix[i][j-1] == matrix[i][j] {
				j--
			} else {
				if matrix[i-1][j] == matrix[i][j] {
					i--
				} else {
					s = append([]Diffable{a[i-1]}, s...)
					i--
					j--
				}
			}
		}
	}
	return lcs, &s
}
