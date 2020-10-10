// +build windows

////////////////////////////////////////////////////////////////////////
//	NOTE:
////////////////////////////////////////////////////////////////////////
package limitutil

////////////////////////////////////////////////////////////////

var MAX_FD uint64 = 1<<64 - 1

////////////////////////////////////////////////////////////////

func GrowToMaxFdLimit() error {
	max, err := GetMaxFdLimit()
	if err != nil {
		return err
	}

	if err = SetCurFdLimit(max); err != nil {
		return err
	}
	return nil
}

func GrowToMaxFdLimitRoot() error {
	max, err := GetMaxFdLimitRoot()
	if err != nil {
		return err
	}

	if err = SetCurFdLimitRoot(max); err != nil {
		return err
	}
	return nil
}

////////////////////////////////////////////////////////////////

func GetCurFdLimit() (uint64, error) {
	return 0, nil
}

func GetMaxFdLimit() (uint64, error) {
	return 0, nil
}

func GetMaxFdLimitRoot() (uint64, error) {
	return MAX_FD, nil
}

/////////////////////////////////////////////////////////////////////////////////

func SetCurFdLimit(fdNum uint64) error {
	return nil
}

func SetCurFdLimitRoot(fdNum uint64) error {
	return nil
}
