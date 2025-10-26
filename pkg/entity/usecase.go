package entity

type UseCase[P any, R any] interface {
	Invoke(param P) (R, error)
	InitService()
}
