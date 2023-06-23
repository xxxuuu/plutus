package app

var services []Service
var notices []Notice

func RegisterService(s Service) {
	services = append(services, s)
}

func RegisterNotice(n Notice) {
	notices = append(notices, n)
}
