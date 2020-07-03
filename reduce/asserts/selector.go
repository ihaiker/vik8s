package asserts

import "github.com/ihaiker/vik8s/reduce/config"

func Selector(d *config.Directive, comsumer func(*config.Directive)) {
	if len(d.Args) == 0 {
		for _, body := range d.Body {
			comsumer(&config.Directive{
				Name: body.Name, Args: body.Args,
				Body: body.Body,
			})
		}
	} else {
		comsumer(&config.Directive{
			Name: d.Args[0], Args: d.Args[1:], Body: d.Body,
		})
	}
}
