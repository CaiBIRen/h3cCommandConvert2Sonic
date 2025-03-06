package sonicmodel

// 定义 VRF 列表中的元素结构体
type Vrf struct {
    VrfName  string `json:"vrf_name,omitempty"`
    // Fallback *bool   `json:"fallback,omitempty"`
    Vni      int    `json:"vni,omitempty"`
}

// 定义 VRF 对象结构体
type VRF struct {
    VRF_LIST []Vrf `json:"VRF_LIST,omitempty"`
}

// 定义顶层结构体 SonicVRF
type SonicVRF struct {
    VRF VRF `json:"VRF,omitempty"`
}

// 定义 JSON 根结构体 Vrfroot
type Vrfroot struct {
    SonicVrf SonicVRF `json:"sonic-vrf:sonic-vrf,omitempty"`
}