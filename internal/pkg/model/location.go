package model

type (
	// Location 定位信息
	Location struct {
		Longitude float64 `json:"longitude"` // 经度
		Latitude  float64 `json:"latitude"`  // 纬度
		Altitude  float64 `json:"altitude"`  // 海拔
		Accuracy  float64 `json:"accuracy"`  // 精度
		Direction float64 `json:"direction"` // 方向
		Speed     float64 `json:"speed"`     // 速度

		Address  string `json:"address"`  // 综合地址
		Country  string `json:"country"`  // 国家
		Province string `json:"province"` // 省
		City     string `json:"city"`     // 城/市
		District string `json:"district"` // 区
		Street   string `json:"street"`   // 街道

		CityId string `json:"cityId"` // 城市编号
	}
)
