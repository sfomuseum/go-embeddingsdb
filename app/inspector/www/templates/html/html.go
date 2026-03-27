package html

import (
	"context"
	"embed"
	"html/template"

	sfom_html "github.com/sfomuseum/go-template/html"
)

//go:embed *.html
var FS embed.FS

func LoadTemplates(ctx context.Context) (*template.Template, error) {
	return sfom_html.LoadTemplates(ctx, FS)
}
