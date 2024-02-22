package maplist

import (
	"bufio"
	_ "embed"
	"fmt"
	"strings"
)

//go:embed maplist.con
var defaultSource []byte

type MapInfo struct {
	Name  string `json:"name"`
	Mode  string `json:"mode"`
	Layer string `json:"layer"`
}

var DefaultMapList = Parse(string(defaultSource))

func Parse(maplistContent string) []MapInfo {
	reader := strings.NewReader(maplistContent)
	scanner := bufio.NewScanner(reader)

	var maps []MapInfo

	for scanner.Scan() {
		text := scanner.Text()

		if !strings.HasPrefix(text, "mapList.append") {
			continue
		}

		parts := strings.Split(text, " ")

		maps = append(maps, MapInfo{
			Name:  parts[1],
			Mode:  parts[2],
			Layer: parts[3],
		})
	}

	return maps
}

func Compose(maps []MapInfo) string {
	builder := strings.Builder{}

	for _, m := range maps {
		builder.WriteString(fmt.Sprintf("mapList.append %s %s %s\n", m.Name, m.Mode, m.Layer))
	}

	return builder.String()
}

func Filter(allMaps []MapInfo, filter MapInfo) []MapInfo {
	var maps []MapInfo

	for _, m := range allMaps {
		if strings.Contains(m.Name, filter.Name) && strings.Contains(m.Mode, filter.Mode) && strings.Contains(m.Layer, filter.Layer) {
			maps = append(maps, m)
		}
	}

	return maps
}
