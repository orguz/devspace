package sync

import (
	"strconv"
	"strings"

	"github.com/juju/errors"
)

// IsDirectory is a constant that can be used to determine whether a file is a folder
const IsDirectory uint64 = 040000

// IsRegularFile is a constant that can be used to determine whether a file is a regular file
const IsRegularFile uint64 = 0100000

// IsSymbolicLink is a constant that can be used to determine whether a file is a symbolic link
const IsSymbolicLink uint64 = 0120000

type fileInformation struct {
	Name  string // %n
	Size  int64  // %s
	Mtime int64  // %Y

	IsSymbolicLink bool // parseHex(%f) & 0120000
	IsDirectory    bool // parseHex(%f) & 040000

	RemoteMode int64 // %a
	RemoteUID  int   // %g
	RemoteGID  int   // %u
}

func getFindCommand(destPath string) string {
	return "mkdir -p '" + destPath + "' && find '" + destPath + "' -exec stat -c \"%n///%s,%Y,%f,%a,%u,%g\" {} + 2>/dev/null && echo -n \"" + EndAck + "\" || echo \"" + ErrorAck + "\"\n"
}

func parseFileInformation(fileline, destPath string) (*fileInformation, error) {
	fileinfo := fileInformation{}

	t := strings.Split(fileline, "///")

	if len(t) != 2 {
		return nil, errors.New("[Downstream] Wrong fileline: " + fileline)
	}

	if len(t[0]) <= len(destPath) {
		return nil, nil
	}

	fileinfo.Name = t[0][len(destPath):]

	t = strings.Split(t[1], ",")

	if len(t) != 6 {
		return nil, errors.New("[Downstream] Wrong fileline: " + fileline)
	}

	size, err := strconv.Atoi(t[0])

	if err != nil {
		return nil, errors.Trace(err)
	}

	fileinfo.Size = int64(size)

	mTime, err := strconv.Atoi(t[1])

	if err != nil {
		return nil, errors.Trace(err)
	}

	fileinfo.Mtime = int64(mTime)

	rawMode, err := strconv.ParseUint(t[2], 16, 32) // Parse hex string into uint64

	if err != nil {
		return nil, errors.Trace(err)
	}

	// We don't sync symbolic links because there are problems on windows
	fileinfo.IsSymbolicLink = (rawMode & IsSymbolicLink) == IsSymbolicLink
	fileinfo.IsDirectory = (rawMode & IsDirectory) == IsDirectory

	mode, err := strconv.ParseInt(t[3], 8, 32)

	if err != nil {
		return nil, errors.Trace(err)
	}

	fileinfo.RemoteMode = mode

	uid, err := strconv.Atoi(t[4])

	if err != nil {
		return nil, errors.Trace(err)
	}

	fileinfo.RemoteUID = uid

	gid, err := strconv.Atoi(t[5])

	if err != nil {
		return nil, errors.Trace(err)
	}

	fileinfo.RemoteGID = gid

	return &fileinfo, nil
}