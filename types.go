package main

type Message struct {
	Id    int
	Views int
}

type Messages []Message

func (m Messages) Less(i, j int) bool {
	return m[i].Views > m[j].Views
}

func (m Messages) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (m Messages) Len() int {
	return len(m)
}
