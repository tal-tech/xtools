// +build !windows

////////////////////////////////////////////////////////////////////////
//	NOTE:	This utils are non-thread safe, non-high-performance
//	utils, use it carefully.
////////////////////////////////////////////////////////////////////////
package limitutil

import (
	"errors"
	"runtime"
	"syscall"
)

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
	var rlimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rlimit)
	if err != nil {
		return 0, err
	}

	return rlimit.Cur, nil
}

func GetMaxFdLimit() (uint64, error) {
	var rlimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rlimit)
	if err != nil {
		return 0, err
	}

	return rlimit.Max, nil
}

func GetMaxFdLimitRoot() (uint64, error) {
	var rlimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rlimit)
	if err != nil {
		return 0, err
	}

	if rlimit.Max > MAX_FD {
		MAX_FD = rlimit.Max
	}
	return MAX_FD, nil
}

/////////////////////////////////////////////////////////////////////////////////

func SetCurFdLimit(fdNum uint64) error {
	if fdNum > MAX_FD {
		return errors.New("fd passed in exceeds the max value(999999)")
	}

	maxLimit, err := GetMaxFdLimit()
	if err != nil {
		return err
	}

	if fdNum > maxLimit {
		return errors.New("fd passed in exceeds the hard limit, you can call SetFdLimitRoot() instead")
	}

	var rlimit syscall.Rlimit
	rlimit.Cur = fdNum
	rlimit.Max = maxLimit
	if runtime.GOOS == "darwin" && rlimit.Cur > 10240 {
		// The max file limit is 10240, even though
		// the max returned by Getrlimit is 1<<63-1.
		// This is OPEN_MAX in sys/syslimits.h.
		rlimit.Cur = 10240
	}

	err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rlimit)
	if err != nil {
		return err
	}
	return nil
}

func SetCurFdLimitRoot(fdNum uint64) error {
	var rlimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rlimit)
	if err != nil {
		return err
	}

	//	increse the hard limit
	if rlimit.Max < fdNum {
		rlimit.Max = fdNum
		rlimit.Cur = fdNum
	}
	//	decrease the hard limit
	if rlimit.Cur > fdNum {
		rlimit.Cur = fdNum
	}
	if runtime.GOOS == "darwin" && rlimit.Cur > 10240 {
		// The max file limit is 10240, even though
		// the max returned by Getrlimit is 1<<63-1.
		// This is OPEN_MAX in sys/syslimits.h.
		rlimit.Cur = 10240
	}

	err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rlimit)
	if err != nil {
		return err
	}
	return nil
}
