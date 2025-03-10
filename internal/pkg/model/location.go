package model

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

type (
	// Location 定位信息
	Location struct {
		// 坐标信息
		Longitude float64 `json:"longitude"` // 经度，东经为正，西经为负
		Latitude  float64 `json:"latitude"`  // 纬度，北纬为正，南纬为负
		Altitude  float64 `json:"altitude"`  // 海拔高度(米)
		Accuracy  float64 `json:"accuracy"`  // 精度(米)
		Direction float64 `json:"direction"` // 方向(度)，正北为0，顺时针方向
		Speed     float64 `json:"speed"`     // 速度(米/秒)

		// 地址信息
		Address  string `json:"address"`  // 完整地址
		Country  string `json:"country"`  // 国家
		Province string `json:"province"` // 省/州
		City     string `json:"city"`     // 城市
		District string `json:"district"` // 区/县
		Street   string `json:"street"`   // 街道名称
		StreetNo string `json:"streetNo"` // 街道号码

		// 标识信息
		CityId      string           `json:"cityId"`           // 城市编号
		PostalCode  string           `json:"postalCode"`       // 邮政编码
		CountryCode string           `json:"countryCode"`      // 国家代码(ISO)
		POI         string           `json:"poi"`              // 兴趣点名称
		Timestamp   int64            `json:"timestamp"`        // 定位时间戳
		Source      string           `json:"source"`           // 数据来源
		Confidence  float64          `json:"confidence"`       // 置信度(0-1)
		System      CoordinateSystem `json:"system,omitempty"` // 坐标系统
	}

	// CoordinateSystem 坐标系统类型
	CoordinateSystem int

	// GeoArea 地理区域
	GeoArea struct {
		Center   *Location   `json:"center"`             // 中心点
		Radius   float64     `json:"radius"`             // 半径(米)
		Name     string      `json:"name"`               // 区域名称
		Vertices []*Location `json:"vertices,omitempty"` // 多边形顶点
	}
)

// 坐标系统常量
const (
	WGS84 CoordinateSystem = iota // 国际标准GPS坐标系
	GCJ02                         // 国测局坐标系(火星坐标系)
	BD09                          // 百度坐标系
)

// NewLocation 创建Location实例
func NewLocation(longitude, latitude float64) *Location {
	return &Location{
		Longitude: longitude,
		Latitude:  latitude,
		Timestamp: time.Now().Unix(),
	}
}

// NewLocationWithAddress 创建带地址的Location实例
func NewLocationWithAddress(longitude, latitude float64, address string) *Location {
	return &Location{
		Longitude: longitude,
		Latitude:  latitude,
		Address:   address,
		Timestamp: time.Now().Unix(),
	}
}

// NewLocationComplete 创建完整地址信息的位置
func NewLocationComplete(longitude, latitude float64, country, province, city, district, street string) *Location {
	loc := NewLocation(longitude, latitude)
	loc.Country = country
	loc.Province = province
	loc.City = city
	loc.District = district
	loc.Street = street
	loc.Address = loc.FormatAddress()
	return loc
}

// IsValid 检查坐标是否有效
func (l *Location) IsValid() bool {
	return l != nil && l.Longitude >= -180 && l.Longitude <= 180 &&
		l.Latitude >= -90 && l.Latitude <= 90
}

// IsEmpty 检查坐标是否为空或无效
func (l *Location) IsEmpty() bool {
	return l == nil || (l.Longitude == 0 && l.Latitude == 0)
}

// FormatAddress 格式化地址
func (l *Location) FormatAddress() string {
	if l.Address != "" {
		return l.Address
	}

	var parts []string
	for _, part := range []string{
		l.Country,
		l.Province,
		l.City,
		l.District,
		l.Street,
		l.StreetNo,
	} {
		if part != "" {
			parts = append(parts, part)
		}
	}

	return strings.Join(parts, " ")
}

// DistanceTo 计算与另一位置的距离(米)
func (l *Location) DistanceTo(other *Location) float64 {
	if l == nil || other == nil {
		return -1
	}

	const earthRadius = 6371000.0 // 地球平均半径，单位米

	lat1 := l.Latitude * math.Pi / 180
	lon1 := l.Longitude * math.Pi / 180
	lat2 := other.Latitude * math.Pi / 180
	lon2 := other.Longitude * math.Pi / 180

	dLat := lat2 - lat1
	dLon := lon2 - lon1

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

// BearingTo 计算到另一位置的方位角(度)
func (l *Location) BearingTo(other *Location) float64 {
	if l == nil || other == nil {
		return 0
	}

	lat1 := l.Latitude * math.Pi / 180
	lon1 := l.Longitude * math.Pi / 180
	lat2 := other.Latitude * math.Pi / 180
	lon2 := other.Longitude * math.Pi / 180

	y := math.Sin(lon2-lon1) * math.Cos(lat2)
	x := math.Cos(lat1)*math.Sin(lat2) - math.Sin(lat1)*math.Cos(lat2)*math.Cos(lon2-lon1)
	bearing := math.Atan2(y, x) * 180 / math.Pi

	// 转换为0-360度范围
	if bearing < 0 {
		bearing += 360
	}
	return bearing
}

// IsWithinDistance 判断是否在指定距离内
func (l *Location) IsWithinDistance(other *Location, distance float64) bool {
	return l.DistanceTo(other) <= distance
}

// Clone 克隆位置信息
func (l *Location) Clone() *Location {
	if l == nil {
		return nil
	}

	clone := *l
	return &clone
}

// String 将位置信息转为字符串
func (l *Location) String() string {
	return fmt.Sprintf("Location{%.6f,%.6f %s}",
		l.Longitude, l.Latitude, l.FormatAddress())
}

// ToMap 转换为Map表示
func (l *Location) ToMap() map[string]interface{} {
	if l == nil {
		return nil
	}

	result := make(map[string]interface{})
	data, err := json.Marshal(l)
	if err != nil {
		return result
	}

	json.Unmarshal(data, &result)
	return result
}

// UpdateFromMap 从Map更新位置信息
func (l *Location) UpdateFromMap(data map[string]interface{}) {
	if l == nil || len(data) == 0 {
		return
	}

	if v, ok := data["longitude"].(float64); ok {
		l.Longitude = v
	}
	if v, ok := data["latitude"].(float64); ok {
		l.Latitude = v
	}
	if v, ok := data["address"].(string); ok {
		l.Address = v
	}
	// 可以添加更多字段更新逻辑
}

// ParseLocation 从经纬度字符串解析位置
func ParseLocation(coordStr string) (*Location, error) {
	coordStr = strings.TrimSpace(coordStr)
	parts := strings.Split(strings.Trim(coordStr, "()[]{}"), ",")

	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid coordinate format")
	}

	lat, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	lng, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)

	if err1 != nil || err2 != nil {
		return nil, fmt.Errorf("invalid coordinate values")
	}

	return NewLocation(lng, lat), nil
}

// IsPointInCircle 判断点是否在圆形区域内
func IsPointInCircle(point *Location, center *Location, radiusMeters float64) bool {
	if point == nil || center == nil || radiusMeters <= 0 {
		return false
	}
	return center.DistanceTo(point) <= radiusMeters
}

// IsPointInPolygon 判断点是否在多边形区域内
func IsPointInPolygon(point *Location, vertices []*Location) bool {
	if point == nil || len(vertices) < 3 {
		return false
	}

	intersections := 0
	verticesCount := len(vertices)

	for i := 0; i < verticesCount; i++ {
		j := (i + 1) % verticesCount

		// 检查射线是否与边相交
		if ((vertices[i].Latitude > point.Latitude) != (vertices[j].Latitude > point.Latitude)) &&
			(point.Longitude < (vertices[j].Longitude-vertices[i].Longitude)*(point.Latitude-vertices[i].Latitude)/
				(vertices[j].Latitude-vertices[i].Latitude)+vertices[i].Longitude) {
			intersections++
		}
	}

	// 奇数次相交表示在多边形内部
	return intersections%2 == 1
}

// NewGeoArea 创建圆形地理区域
func NewGeoArea(center *Location, radiusMeters float64, name string) *GeoArea {
	return &GeoArea{
		Center: center,
		Radius: radiusMeters,
		Name:   name,
	}
}

// ContainsPoint 检查区域是否包含指定点
func (g *GeoArea) ContainsPoint(point *Location) bool {
	if g == nil || point == nil {
		return false
	}

	if len(g.Vertices) >= 3 {
		// 多边形区域检查
		return IsPointInPolygon(point, g.Vertices)
	}

	// 圆形区域检查
	return g.Center.DistanceTo(point) <= g.Radius
}
