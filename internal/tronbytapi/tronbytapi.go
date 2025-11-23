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

type PushPayload struct {
	DeviceID       string `json:"deviceID"`
	Image          string `json:"image"`
	InstallationID string `json:"installationID"`
	Background     bool   `json:"background"`
}
