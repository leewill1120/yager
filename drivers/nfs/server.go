package nfs

var (
	defaultLocalPath = "/mnt/yager/nfs"
)

type Server struct {
	LocalPath string
}

func NewServer(localPath string) *Server {
	if 0 == len(localPath) {
		localPath = defaultLocalPath
	}
	return &Server{
		LocalPath: localPath,
	}
}
