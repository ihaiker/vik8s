package tools

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ihaiker/vik8s/config"
	"github.com/ihaiker/vik8s/libs/logs"
	"github.com/ihaiker/vik8s/libs/utils"
)

type input = func(context.Context, *schema.ResourceData, *config.Configuration) error
type output = func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics

func Logger(method string, fn input) output {
	return func(ctx context.Context, data *schema.ResourceData, i interface{}) (dd diag.Diagnostics) {
		defer utils.Catch(func(e error) {
			logs.Info(utils.Stack())
			dd = diag.FromErr(e)
		})
		if err := fn(ctx, data, i.(*config.Configuration)); err != nil {
			return diag.FromErr(err)
		}
		if err := i.(*config.Configuration).Write(); err != nil {
			return diag.FromErr(err)
		}
		return
	}
}

func SetValue(str []string) *schema.Set {
	ins := make([]interface{}, len(str))
	for i, s := range str {
		ins[i] = s
	}
	return schema.NewSet(schema.HashString, ins)
}

func GetDataSetString(p interface{}) []string {
	if p == nil {
		return nil
	}
	items := p.(*schema.Set).List()
	strs := make([]string, len(items))
	for i, item := range items {
		strs[i] = item.(string)
	}
	return strs
}

func Length(p interface{}) int {
	if p == nil {
		return 0
	}
	if list, match := p.([]interface{}); match {
		return len(list)
	}
	if set, match := p.(*schema.Set); match {
		return set.Len()
	}
	return 0
}

func SetState(p interface{}, data *schema.ResourceData) error {
	for k, v := range p.([]interface{})[0].(map[string]interface{}) {
		if err := data.Set(k, v); err != nil {
			return err
		}
	}
	return nil
}

func Id(prefix string, i interface{}) string {
	bs, _ := json.Marshal(i)
	return fmt.Sprintf("%s-%x", prefix, md5.Sum(bs))
}
