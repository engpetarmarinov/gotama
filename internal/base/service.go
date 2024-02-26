package base

type Service interface {
	Run()
	Shutdown() error
}
