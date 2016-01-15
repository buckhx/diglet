package mbt

type Properties struct {
	props map[string]interface{}
}

func PropertiesFromMap(props map[string]interface{}) *Properties {
	return &Properties{props}
}

func (p *Properties) Set(key string, value interface{}) {
	p.props[key] = value
}

func (p *Properties) GetString(key string) string {
	return p.props[key].(string)
}

func (p *Properties) GetFloat(key string) float32 {
	return p.props[key].(float32)
}

func (p *Properties) GetDouble(key string) float64 {
	return p.props[key].(float64)
}

func (p *Properties) GetInt(key string) int64 {
	return p.props[key].(int64)
}

func (p *Properties) GetUint(key string) uint64 {
	return p.props[key].(uint64)
}

func (p *Properties) GetBool(key string) bool {
	return p.props[key].(bool)
}

func (p *Properties) Get(key string) interface{} {
	return p.props[key]
}
