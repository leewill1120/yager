package volume

//status of volume
const (
	INIT = iota
	MOUNTED
	UNMOUNTED
	ABNORMAL
)

type CommonVolume struct {
	Name       string
	Status     int
	Type       string //iscsi nfs cifs
	MountPoint string
}

type Volume interface {
	Mount() error
	Umount() error
	Attribute() map[string]interface{}
}
