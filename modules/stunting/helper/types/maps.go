package types

import "fmt"

type Mapper struct {
	Map     map[any]any
	Label   string
	Default any
}

var pengukuran = map[any]any{
	1: "Tidur",
	2: "Berdiri",
}

var (
	Pengukuran = Mapper{
		Map:     pengukuran,
		Label:   "Pengukuran",
		Default: nil,
	}
)

func stringPtr(s string) *string {
	return &s
}

func (m Mapper) ToLabel(key any) *string {
	if val, ok := m.Map[key]; ok && val != nil {
		str := fmt.Sprint(val)
		return &str
	}

	if m.Default != nil {
		str := fmt.Sprint(m.Default)
		return &str
	}

	return nil
}

func (m Mapper) ToKey(key any, def *string) *string {
	for k, v := range m.Map {
		if v == key {
			strKey := fmt.Sprint(k)
			return &strKey
		}
	}

	if def != nil {
		strDefault := fmt.Sprint(def)
		return &strDefault
	}

	return nil
}
