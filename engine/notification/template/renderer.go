package template

import (
	"fmt"
	"regexp"
)

type renderer struct{}

func NewRenderer() *renderer {
	return &renderer{}
}

var placeholderRegex = regexp.MustCompile(`\{\{(\w+)\}\}`)

func (r *renderer) Render(template string, data map[string]any) (string, error) {

	matches := placeholderRegex.FindAllStringSubmatch(template, -1)

	for _,match:=range matches{
		key:=match[1]

		if _,ok:=data[key]; !ok{
				return "", fmt.Errorf("missing template variable: %q", key)
		}
	}

	result:=placeholderRegex.ReplaceAllStringFunc(template,func(match string) string {
		key:=match[2:len(match)-2]
		return fmt.Sprintf("%v", data[key])
	})

	return result, nil

}
