package tronbytapi

type Devices struct {
	Devices []Device `json:"devices"`
}

type Device struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
}

type Installations struct {
	Installations []Installation `json:"installations"`
}

type Installation struct {
	Id    string `json:"id"`
	AppId string `json:"appID"`
}
