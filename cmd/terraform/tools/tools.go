package tools

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ihaiker/vik8s/cmd/terraform/configure"
	"github.com/ihaiker/vik8s/libs/utils"
	"log"
	"strings"
)

type input = func(context.Context, *schema.ResourceData, *configure.MemStorage) diag.Diagnostics
type output = func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics

func Safe(method string, fn input) output {
	return func(ctx context.Context, data *schema.ResourceData, i interface{}) (dd diag.Diagnostics) {
		defer utils.Catch(func(err error) {
			dd = diag.FromErr(err)
		})
		log.Println(strings.Repeat("<", 30), method, strings.Repeat("<", 30))
		defer func() { log.Println(strings.Repeat(">", 30), method, strings.Repeat("<", 30)) }()
		return fn(ctx, data, i.(*configure.MemStorage))
	}
}

func SetValue(str []string) *schema.Set {
	ins := make([]interface{}, len(str))
	for i, s := range str {
		ins[i] = s
	}
	return schema.NewSet(schema.HashString, ins)
}

func StringValue(str string) *string {
	if str == "" {
		return nil
	}
	return &str
}
