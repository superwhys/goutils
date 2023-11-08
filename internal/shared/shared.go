package shared

var (
	PtrServiceName *string
	PtrConsulAddr  *string
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
