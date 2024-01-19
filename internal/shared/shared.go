package shared

var (
	PtrServiceName *string
	PtrConsulAddr  *string
	LogToFile      *bool
	LogFileName    *string
	LogMaxSize     *int
	LogMaxBackup   *int
	LogMaxAge      *int
	LogCompress    *bool
)

func GetServiceName() string {
	if PtrServiceName == nil {
		return ""
	}
	return *PtrServiceName
}

func GetConsulAddress() string {
	if PtrConsulAddr == nil {
		return ""
	}
	return *PtrConsulAddr
}

func IsLogToFile() bool {
	if LogToFile == nil {
		return false
	}
	return *LogToFile
}

func GetLogFileName() string {
	if LogFileName == nil {
		return "runlog.log"
	}
	return *LogFileName
}

func GetLogMaxSize() int {
	if LogMaxSize == nil {
		return 10
	}
	return *LogMaxSize
}

func GetLogMaxBackup() int {
	if LogMaxBackup == nil {
		return 3
	}
	return *LogMaxBackup
}

func GetLogMaxAge() int {
	if LogMaxAge == nil {
		return 30
	}
	return *LogMaxAge
}

func IsLogCompress() bool {
	if LogCompress == nil {
		return false
	}
	return *LogCompress
}
