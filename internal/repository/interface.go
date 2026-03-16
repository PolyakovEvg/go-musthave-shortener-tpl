package repository

type Repository interface {
	Save(string) (string, error)
	Get(string) (string, bool)
}
