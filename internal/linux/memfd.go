package linux

import (
	"os"

	"golang.org/x/sys/unix"

	"github.com/lxc/incus/v6/shared/revert"
)

// CreateMemfd creates a new memfd for the provided byte slice.
func CreateMemfd(content []byte) (*os.File, error) {
	reverter := revert.New()
	defer reverter.Fail()

	// Create the memfd.
	fd, err := unix.MemfdCreate("memfd", unix.MFD_CLOEXEC)
	if err != nil {
		return nil, err
	}

	reverter.Add(func() { _ = unix.Close(fd) })

	// Set its size.
	err = unix.Ftruncate(fd, int64(len(content)))
	if err != nil {
		return nil, err
	}

	// Prepare the storage.
	data, err := unix.Mmap(fd, 0, len(content), unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
	if err != nil {
		return nil, err
	}

	// Write the content.
	copy(data, content)

	// Cleanup.
	err = unix.Munmap(data)
	if err != nil {
		return nil, err
	}

	reverter.Success()

	return os.NewFile(uintptr(fd), "memfd"), nil
}
