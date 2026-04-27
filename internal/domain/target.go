package domain

type Target struct {
	URL             string
	LastTimeChecked string
	LastCode        int
	Interval        int
}
