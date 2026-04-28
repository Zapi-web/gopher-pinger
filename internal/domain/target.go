package domain

type Target struct {
	ID              string
	URL             string
	LastTimeChecked string
	LastCode        int
	Interval        int
}
