package models

type Content struct {
	Type   int         `json:"type"`
	Data   string      `json:"data"`
	Detail Detail      `json:"detail,omitempty"`
	Desc   string      `json:"desc,omitempty"`
	User   ContentUser `json:"user"`
	Img    []byte      `json:"img,omitempty"`
	Voice  []byte      `json:"voice,omitempty"`
	Other  interface{} `json:"other,omitempty"`
}

type Detail struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}
