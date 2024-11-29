package domain

type User struct {
	ID       uint64
	Login    string
	Password string
	Balance  *Balance
}
