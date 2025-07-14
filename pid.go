package pid

type PIDFetcher interface {
	GetPID(packageName string) (int, error)
}

type PidofFetcher struct{

}

func (p PidofFetcher) GetPID(packageName string) (int, error) {

}


